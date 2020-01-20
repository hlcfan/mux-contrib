package recovery

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// Logger is used to print logs
type Logger interface {
	Println(...interface{})
}

// Middleware is the rescue to panic
type Middleware struct {
	logger Logger
}

// OverrideLogger creates Middleware
func (m *Middleware) OverrideLogger(logger Logger) {
	m.logger = logger
}

// Middleware recovers from panic
func (m *Middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next = handlers.RecoveryHandler(
			handlers.PrintRecoveryStack(true),
			handlers.RecoveryLogger(m.logger),
		)(next)
		next.ServeHTTP(w, r)
	})
}
