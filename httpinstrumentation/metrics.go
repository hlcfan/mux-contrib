package httpinstrumentation

// import (
// 	"fmt"
// 	"strconv"

// 	"github.com/prometheus/client_golang/prometheus"
// )

// var (
// 	// LatencyHist represents request duration
// 	latencyHist = prometheus.NewHistogramVec(
// 		prometheus.HistogramOpts{
// 			Name:    "ls_http_request_duration_seconds",
// 			Help:    "A histogram of latencies for requests",
// 			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5}, // Measure latencies in buckets between 1ms and 5s
// 		},
// 		[]string{"path", "code"},
// 	)
// )

// // Metrics handles the app's metrics
// type Metrics struct {
// 	LatencyHist *prometheus.HistogramVec
// }

// // NewMetrics creates metric instance
// func NewMetrics() *Metrics {
// 	return &Metrics{
// 		LatencyHist: latencyHist,
// 	}
// }

// func init() {
// 	// Metrics have to be registered to be exposed:
// 	prometheus.MustRegister(latencyHist)
// }

// // ReportLatency sends duration by route name
// func (m *Metrics) ReportLatency(routeName string, statusCode int, duration float64) {
// 	fmt.Println("===Report: ", routeName)
// 	m.LatencyHist.WithLabelValues(routeName, strconv.Itoa(statusCode)).Observe(duration)
// }
