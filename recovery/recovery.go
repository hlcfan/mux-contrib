package middleware

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// Logger is used to print logs
type Logger interface {
	Println(...interface{})
}

// RecoveryMiddleware is the rescue to panic
type RecoveryMiddleware struct {
	logger Logger
}

// OverrideLogger creates RecoveryMiddleware
func (m *RecoveryMiddleware) OverrideLogger(logger Logger) {
	m.logger = logger
}

// Middleware recovers from panic
func (m *RecoveryMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next = handlers.RecoveryHandler(
			handlers.PrintRecoveryStack(true),
			handlers.RecoveryLogger(m.logger),
		)(next)
		next.ServeHTTP(w, r)
	})
}
