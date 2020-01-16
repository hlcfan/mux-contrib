package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var urlBlacklist = map[string]struct{}{
	"/metrics": struct{}{},
}

// apacheFormatPattern follows Apache CLF (combined)
// Extend the format here to add the Request ID
const apacheFormatPattern = "%s - - [%s] \"%s %d %d\" %f \"%s\" \"%s\" \"%s\"\n"

// HTTPInstrumentationMiddleware is the middleware to collect metrics
type HTTPInstrumentationMiddleware struct {
	options Options
	router  *mux.Router
	hooks   []callback
	output  io.Writer
}

// Options is used to configure this middleware
// Logging is enabled by default
type Options struct {
	Logging bool
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

// customResponseWriter is used for getting the response status
type customResponseWriter struct {
	http.ResponseWriter
	status int
	length int
}

// NewHTTPInstrumentationMiddleware creates instrumentation middleware
func NewHTTPInstrumentationMiddleware(router *mux.Router) *HTTPInstrumentationMiddleware {
	m := &HTTPInstrumentationMiddleware{
		output: os.Stdout,
		options: Options{
			Logging: true,
		},
		router: router,
	}

	m.RegisterHook(m.loggingProcessor)
	return m
}

// SetOutput specifies the output, follows io.Writer
func (middleware *HTTPInstrumentationMiddleware) SetOutput(output io.Writer) {
	middleware.output = output
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

// DisableLogging disables the logging
func (middleware *HTTPInstrumentationMiddleware) DisableLogging() {
	middleware.options.Logging = false
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

func (middleware *HTTPInstrumentationMiddleware) loggingProcessor(r *InstrumentationRecord) {
	if middleware.options.Logging {
		timeFormatted := r.Timestamp.Format("02/Jan/2006 03:04:05")
		requestLine := fmt.Sprintf("%s %s %s", r.Method, r.URI, r.Protocol)
		fmt.Fprintf(middleware.output, apacheFormatPattern, r.IPAddr, timeFormatted, requestLine, r.Status, r.ResponseBytes, r.ElapsedTime.Seconds(), r.Referer, r.UserAgent, r.RequestID)
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
