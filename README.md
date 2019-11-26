# mux-contrib

## HTTP Instrumentation

It'll report `route name`, `status code`, `duration`.

### Usage

``` go
  router := mux.NewRouter().StrictSlash(true)

  metricReporter := metrics.NewMetrics()
  // `metricReport` should implement `ReportLatency(routeName string, method string, statusCode int, duration float64)`
  mw := middleware.NewHTTPInstrumentationMiddleware(router, metricReporter)

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
