# mux-contrib

## Installation

### Go module
```
go get -u github.com/hlcfan/mux-contrib/middleware
```

### Dep
```
dep ensure -add github.com/hlcfan/mux-contrib/middleware
```


## HTTP Instrumentation

This middleware will throw/expose the HTTP request & response info. In order to retrive the info, you will need to register processors to process the data. Once processors being registered, it's able to consume the data. The data is defined as following

```
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
```

Use the data as you want, like sending metrics or logging.

By default, logging is enabled and follows Apache CLF (combined) format, plus a request ID, can be used for tracking requests. It looks like
```
127.0.0.1:58004 - - [16/Jan/2020 09:35:08] "GET /v1/books/1" 0.013809 "" "PostmanRuntime/7.21.0" "0xashi.local/CHzHHzy75q-000001"
```

### Usage

``` go
router := mux.NewRouter().StrictSlash(true)

mw := middleware.NewHTTPInstrumentationMiddleware(router)

// Register processors
mw.RegisterHook(func(record *middleware.InstrumentationRecord) {
    fmt.Println("Instrumentation data: %#v", record)
})

// Disable logging
mw.DisableLogging()


// Set output, can be any io.Writer
mw.SetOutput(os.Stdout)

// Apply middleware
h = mw.Middleware(router)

// OR
router.Use(metricMiddleware.Middleware)
```

## Panic Recovery

It'll recover from panic and log the stacktrace.

### Usage

``` go
mw := middleware.RecoveryMiddleware{}

// Override default logger
// logger should implement `Println(...interface{})`
logger := log.New(os.Stderr, "", 0)
mw.OverrideLogger(logger)

// Apply middleware
h = mw.Middleware(router)

// OR
router.Use(metricMiddleware.Middleware)
```
