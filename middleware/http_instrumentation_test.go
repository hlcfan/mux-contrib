package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/hlcfan/mux-contrib/middleware"
	"github.com/stretchr/testify/assert"
)

const (
	response = "<html><body>Hello World!</body></html>"
	referer  = "http://example.com/"
	ua       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36"
)

func TestMiddleware(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Referer", referer)
		r.Header.Set("User-Agent", ua)
		io.WriteString(w, response)
	}
	router.HandleFunc("/test", handler).Methods("GET").Name("TestPath")

	mw := middleware.NewHTTPInstrumentationMiddleware(router)
	mw.RegisterHook(func(record *middleware.InstrumentationRecord) {
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
}

func TestMiddlewareSkipRoute(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}
	router.HandleFunc("/metrics", handler).Methods("GET").Name("TestPath")

	mw := middleware.NewHTTPInstrumentationMiddleware(router)

	called := 0
	fn := func(record *middleware.InstrumentationRecord) {
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

	mw := middleware.NewHTTPInstrumentationMiddleware(router)
	middleware.BlacklistURL("/path-to-skip")
	called := 0
	fn := func(record *middleware.InstrumentationRecord) {
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
