package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/hlcfan/mux-contrib/middleware"
)

const response = "<html><body>Hello World!</body></html>"

func TestMiddleware(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}
	router.HandleFunc("/test", handler).Methods("GET").Name("TestPath")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockMetricReporter(ctrl)
	mw := middleware.NewHTTPInstrumentationMiddleware(router, m)

	handlerToTest := mw.Middleware(http.HandlerFunc(handler))
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	m.
		EXPECT().
		ReportLatency(gomock.Eq("TestPath"), gomock.Eq(200), gomock.Any())

	handlerToTest.ServeHTTP(w, req)
	// Wait for goroutine to execute
	time.Sleep(100 * time.Millisecond)
}

func TestMiddlewareSkipRoute(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}
	router.HandleFunc("/metrics", handler).Methods("GET").Name("TestPath")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockMetricReporter(ctrl)
	mw := middleware.NewHTTPInstrumentationMiddleware(router, m)

	handlerToTest := mw.Middleware(http.HandlerFunc(handler))
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	m.
		EXPECT().
		ReportLatency(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	handlerToTest.ServeHTTP(w, req)
}

func TestAddBlacklistURL(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}
	router.HandleFunc("/path-to-skip", handler).Methods("GET").Name("TestPath")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockMetricReporter(ctrl)
	mw := middleware.NewHTTPInstrumentationMiddleware(router, m)
	middleware.BlacklistURL("/path-to-skip")

	handlerToTest := mw.Middleware(http.HandlerFunc(handler))
	req := httptest.NewRequest("GET", "/path-to-skip", nil)
	w := httptest.NewRecorder()

	m.
		EXPECT().
		ReportLatency(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	handlerToTest.ServeHTTP(w, req)
}
