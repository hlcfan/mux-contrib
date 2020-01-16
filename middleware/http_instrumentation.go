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
	router *mux.Router
	hooks  []callback
}

// callback consumes the instrumentation data
type callback func(*InstrumentationRecord)

// InstrumentationRecord is the data of request/response info
type InstrumentationRecord struct {
	RouteName     string
	IPAddr        string
	Timestamp     time.Time
	Method        string
	URI           string
	Protocol      string
	Referer       string
	UserAgent     string
	RequestID     string
	Status        int
	ResponseBytes int
	ElapsedTime   time.Duration
}

// NewHTTPInstrumentationMiddleware creates instrumentation middleware
func NewHTTPInstrumentationMiddleware(router *mux.Router) *HTTPInstrumentationMiddleware {
	return &HTTPInstrumentationMiddleware{
		router: router,
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
		sw := &customResponseWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)

		defer func() {
			go func(req *http.Request, sw *customResponseWriter) {
				var match mux.RouteMatch
				if middleware.router.Match(r, &match) && match.Route != nil {
					var routeName string
					routeName = match.Route.GetName()
					if len(routeName) == 0 {
						if r, err := match.Route.GetPathTemplate(); err == nil {
							routeName = r
						}
					}

					record := &InstrumentationRecord{
						RouteName:     routeName,
						IPAddr:        req.RemoteAddr,
						Timestamp:     time.Now().UTC(),
						Method:        req.Method,
						URI:           req.RequestURI,
						Protocol:      req.Proto,
						Referer:       req.Referer(),
						UserAgent:     req.UserAgent(),
						Status:        sw.status,
						ElapsedTime:   time.Since(now),
						ResponseBytes: sw.length,
						RequestID:     requestID(),
					}

					middleware.processHooks(record)
				}
			}(r, sw)
		}()
	})
}

// RegisterHook registers processors
func (middleware *HTTPInstrumentationMiddleware) RegisterHook(callbacks ...callback) {
	for _, cb := range callbacks {
		middleware.hooks = append(middleware.hooks, cb)
	}
}

// processHooks processes hooks one by one
func (middleware *HTTPInstrumentationMiddleware) processHooks(instrumentationRecord *InstrumentationRecord) {
	for _, cb := range middleware.hooks {
		cb(instrumentationRecord)
	}
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
