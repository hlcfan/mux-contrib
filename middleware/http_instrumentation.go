package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var urlBlacklist = map[string]struct{}{
	"/metrics": struct{}{},
}

// HTTPInstrumentationMiddleware is the middleware to collect metrics
type HTTPInstrumentationMiddleware struct {
	Router  *mux.Router
	Metrics MetricReporter
}

// MetricReporter defines what kinds of metrics it supports
// go:generate mockgen -source=http_instrumentation.go -destination=http_instrumentation_test.go -package=middleware
type MetricReporter interface {
	ReportLatency(routeName string, statusCode int, duration float64)
}

// NewHTTPInstrumentationMiddleware creates new metric middleware
func NewHTTPInstrumentationMiddleware(router *mux.Router, metricReporter MetricReporter) *HTTPInstrumentationMiddleware {
	return &HTTPInstrumentationMiddleware{
		Router:  router,
		Metrics: metricReporter,
	}
}

// customResponseWriter is used for getting the response status
type customResponseWriter struct {
	http.ResponseWriter
	status int
	length int
}

// Middleware wraps the logic for collecting metrics
func (middleware *HTTPInstrumentationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if urlBlacklisted(r.RequestURI) {
			next.ServeHTTP(w, r)
			return
		}
		now := time.Now().UTC()
		sw := customResponseWriter{ResponseWriter: w}
		next.ServeHTTP(&sw, r)
		httpDuration := time.Since(now)
		defer func() {
			go func(duration time.Duration, statusCode int) {
				var match mux.RouteMatch
				if middleware.Router.Match(r, &match) && match.Route != nil {
					var routeName string
					routeName = match.Route.GetName()
					if len(routeName) == 0 {
						if r, err := match.Route.GetPathTemplate(); err == nil {
							routeName = r
						}
					}
					middleware.Metrics.ReportLatency(routeName, statusCode, duration.Seconds())
				}
			}(httpDuration, sw.status)
		}()
	})
}

// BlacklistURL addes urls to blacklist
func BlacklistURL(urls ...string) {
	for _, u := range urls {
		urlBlacklist[u] = struct{}{}
	}
}

func urlBlacklisted(url string) bool {
	_, ok := urlBlacklist[url]
	return ok
}

// WriteHeader implements the interface
func (w *customResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Write implements the interface
func (w *customResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}
