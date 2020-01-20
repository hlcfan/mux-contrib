package recovery_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hlcfan/mux-contrib/recovery"
)

var router = mux.NewRouter().StrictSlash(true)
var handler = func(w http.ResponseWriter, r *http.Request) {
	panic("Unexpected error!")
}

func TestRecoveryMiddleware(t *testing.T) {
	mw := recovery.Middleware{}
	handlerToTest := mw.Middleware(http.HandlerFunc(handler))
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handlerToTest.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != 500 {
		t.Fatalf("%d != %d", resp.StatusCode, 500)
	}
}

func TestSkipsRecoveryMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic", r)
		}
	}()
	http.HandlerFunc(handler).ServeHTTP(w, req)

	log.Fatal("No panic")
}
