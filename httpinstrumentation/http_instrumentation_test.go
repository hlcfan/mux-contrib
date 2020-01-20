package httpinstrumentation_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/hlcfan/mux-contrib/httpinstrumentation"
	"github.com/stretchr/testify/assert"
)

const (
	response = "<html><body>Hello World!</body></html>"
	referer  = "http://example.com/"
	ua       = "test-ua"
)

func TestMiddleware(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Referer", referer)
		r.Header.Set("User-Agent", ua)
		r.Header.Set("Referer", "http://example.com/")
		r.RemoteAddr = "fake-local-addr"
		io.WriteString(w, response)
	}
	router.HandleFunc("/test", handler).Methods("GET").Name("TestPath")

	var buf bytes.Buffer
	mw := httpinstrumentation.NewMiddleware(router, httpinstrumentation.Output(&buf))
	mw.RegisterHook(func(record *httpinstrumentation.Record) {
		assert.Equal(t, record.RouteName, "TestPath")
		assert.Equal(t, record.Method, "GET")
		assert.Equal(t, record.Protocol, "HTTP/1.1")
		assert.Equal(t, record.Referer, "http://example.com/")
		assert.Equal(t, record.Status, 200)
		assert.Equal(t, record.ResponseBytes, 38)
		if record.IPAddr == "" {
			assert.Fail(t, "Invalid IP Addr")
		}
		if record.ElapsedTime < time.Duration(0) {
			assert.Fail(t, "Invalid ElapsedTime")
		}

		if record.ElapsedTime > time.Duration(1*time.Millisecond) {
			assert.Fail(t, "Invalid ElapsedTime")
		}
	})

	handlerToTest := mw.Middleware(http.HandlerFunc(handler))
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handlerToTest.ServeHTTP(w, req)
	time.Sleep(5 * time.Millisecond)
	assert.Regexp(t, `fake-local-addr\s-\s-\s\[\d{1,2}\/\w{3}\/\d{4}\s\d{1,2}:\d{1,2}:\d{1,2}\]\s"GET\s\/test\sHTTP\/1.1\s200\s38"\s\d{1,2}\.?\d*µs\s"http:\/\/example.com\/"\s"test-ua"\s".+"`, buf.String())
}

func TestMiddlewareSkipRoute(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}
	router.HandleFunc("/metrics", handler).Methods("GET").Name("TestPath")

	mw := httpinstrumentation.NewMiddleware(router)

	called := 0
	fn := func(record *httpinstrumentation.Record) {
		called++
	}

	mw.RegisterHook(fn)
	handlerToTest := mw.Middleware(http.HandlerFunc(handler))
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handlerToTest.ServeHTTP(w, req)
	time.Sleep(5 * time.Millisecond)

	assert.Equal(t, called, 0)
}

func TestAddBlacklistURL(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}
	router.HandleFunc("/path-to-skip", handler).Methods("GET").Name("TestPath")

	mw := httpinstrumentation.NewMiddleware(router)
	httpinstrumentation.BlacklistURL("/path-to-skip")
	called := 0
	fn := func(record *httpinstrumentation.Record) {
		called++
	}

	mw.RegisterHook(fn)

	handlerToTest := mw.Middleware(http.HandlerFunc(handler))
	req := httptest.NewRequest("GET", "/path-to-skip", nil)
	w := httptest.NewRecorder()
	handlerToTest.ServeHTTP(w, req)
	time.Sleep(5 * time.Millisecond)

	assert.Equal(t, called, 0)
}

func TestDisableLogging(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}
	router.HandleFunc("/test", handler).Methods("GET").Name("TestPath")

	var buf bytes.Buffer
	mw := httpinstrumentation.NewMiddleware(router, httpinstrumentation.Output(&buf), httpinstrumentation.DisableLogging())

	handlerToTest := mw.Middleware(http.HandlerFunc(handler))
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handlerToTest.ServeHTTP(w, req)
	time.Sleep(5 * time.Millisecond)

	assert.Len(t, buf.String(), 0)
}
