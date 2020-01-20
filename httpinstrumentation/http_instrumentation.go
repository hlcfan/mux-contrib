package httpinstrumentation

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var urlBlacklist = map[string]struct{}{
	"/metrics": struct{}{},
	"/ping":    struct{}{},
	"/health":  struct{}{},
}

// apacheFormatPattern follows Apache CLF (combined)
// Extend the format here to add the Request ID
const apacheFormatPattern = "%s - - [%s] \"%s %d %d\" %s \"%s\" \"%s\" \"%s\"\n"

// Middleware is the middleware to collect metrics
type Middleware struct {
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
type callback func(*Record)

// Record is the data of request/response info
type Record struct {
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

// FuncOption is the option for middleware
type FuncOption func(*Middleware)

// NewMiddleware creates instrumentation middleware
func NewMiddleware(router *mux.Router, options ...FuncOption) *Middleware {
	m := &Middleware{
		output: os.Stdout,
		options: Options{
			Logging: true,
		},
		router: router,
	}

	for _, opt := range options {
		opt(m)
	}

	m.RegisterHook(m.loggingProcessor)
	return m
}

// Output specifies the output, follows io.Writer
func Output(output io.Writer) FuncOption {
	return func(mw *Middleware) {
		mw.output = output
	}
}

// DisableLogging disables the logging
func DisableLogging() FuncOption {
	return func(mw *Middleware) {
		mw.options.Logging = false
	}
}

// Middleware wraps the logic for collecting metrics
func (middleware *Middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if urlBlacklisted(r.RequestURI) {
			next.ServeHTTP(w, r)
			return
		}

		now := time.Now().UTC()
		sw := &customResponseWriter{ResponseWriter: w}
		requestID := requestID()
		ctx := context.WithValue(r.Context(), "request-id", requestID)
		r = r.WithContext(ctx)
		next.ServeHTTP(sw, r)
		finishTime := time.Now().UTC()

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

					record := &Record{
						RouteName:     routeName,
						IPAddr:        req.RemoteAddr,
						Timestamp:     finishTime,
						Method:        req.Method,
						URI:           req.RequestURI,
						Protocol:      req.Proto,
						Referer:       req.Referer(),
						UserAgent:     req.UserAgent(),
						Status:        sw.status,
						ElapsedTime:   finishTime.Sub(now),
						ResponseBytes: sw.length,
						RequestID:     requestID,
					}

					middleware.processHooks(record)
				}
			}(r, sw)
		}()
	})
}

// RegisterHook registers processors
func (middleware *Middleware) RegisterHook(callbacks ...callback) {
	for _, cb := range callbacks {
		middleware.hooks = append(middleware.hooks, cb)
	}
}

// processHooks processes hooks one by one
func (middleware *Middleware) processHooks(instrumentationRecord *Record) {
	for _, cb := range middleware.hooks {
		cb(instrumentationRecord)
	}
}

func (middleware *Middleware) loggingProcessor(r *Record) {
	if middleware.options.Logging {
		timeFormatted := r.Timestamp.Format("02/Jan/2006 03:04:05")
		requestLine := fmt.Sprintf("%s %s %s", r.Method, r.URI, r.Protocol)
		fmt.Fprintf(middleware.output, apacheFormatPattern, r.IPAddr, timeFormatted, requestLine, r.Status, r.ResponseBytes, r.ElapsedTime, r.Referer, r.UserAgent, r.RequestID)
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
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}
