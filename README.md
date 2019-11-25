# mux-contrib

## HTTP Instrumentation

It'll report `route name`, `status code`, `duration`.

### Usage

``` go
  router := mux.NewRouter().StrictSlash(true)

  metricReporter := metrics.NewMetrics()
  mw := middleware.NewHTTPInstrumentationMiddleware(router, metricReporter)

  // Apply middleware
  h = mw.Middleware(router)

  // OR
  router.Use(metricMiddleware.Middleware)
```
