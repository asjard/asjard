package collectors

import (
	"github.com/asjard/asjard/core/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// APIRequestCounter tracks the volume of incoming or outgoing requests.
type APIRequestCounter struct {
	counter *prometheus.CounterVec
}

// NewAPIRequestCounter initializes the 'api_requests_total' metric.
// It tracks requests partitioned by HTTP/gRPC status code, endpoint name, and protocol.
func NewAPIRequestCounter() *APIRequestCounter {
	return &APIRequestCounter{
		counter: metrics.RegisterCounter("api_requests_total",
			"The total number of handled requests",
			[]string{"code", "api", "protocol"}),
	}
}

// Inc increments the request counter for the specified labels.
func (a APIRequestCounter) Inc(code, api, protocol string) {
	if a.counter != nil {
		a.counter.With(map[string]string{
			"code":     code,
			"api":      api,
			"protocol": protocol,
		}).Inc()
	}
}

// APIRequestLatency measures how long requests take to process.
type APIRequestLatency struct {
	latency *prometheus.HistogramVec
}

// NewAPIRequestLatency initializes the 'api_requests_latency_seconds' metric.
// It uses default Prometheus buckets to categorize request durations.
func NewAPIRequestLatency() *APIRequestLatency {
	return &APIRequestLatency{
		latency: metrics.RegisterHistogram("api_requests_latency_seconds",
			"The duration of handled requests",
			[]string{"api", "protocol"},
			prometheus.DefBuckets),
	}
}

// Observe records a new latency measurement for the specified API and protocol.
func (a APIRequestLatency) Observe(api, protocol string, value float64) {
	if a.latency != nil {
		a.latency.With(map[string]string{
			"api":      api,
			"protocol": protocol,
		}).Observe(value)
	}
}

const (
	// Binary unit constants for byte calculations.
	_           = iota
	bKB float64 = 1 << (10 * iota) // 1024 bytes
	bMB                            // 1048576 bytes
)

// sizeBuckets defines the distribution ranges for measuring request/response sizes.
// Ranging from 1KB up to 10MB.
var sizeBuckets = []float64{1.0 * bKB, 2.0 * bKB, 5.0 * bKB, 10.0 * bKB, 100 * bKB, 500 * bKB, 1.0 * bMB, 2.5 * bMB, 5.0 * bMB, 10.0 * bMB}

// APIResponseSize monitors the size of the data being sent back to clients.
type APIResponseSize struct {
	size *prometheus.HistogramVec
}

func NewAPIResponseSize() *APIResponseSize {
	return &APIResponseSize{
		size: metrics.RegisterHistogram("api_response_size_bytes",
			"the response sizes in bytes",
			[]string{"api", "protocol"},
			sizeBuckets),
	}
}

// Observe records the response size in bytes.
func (a APIResponseSize) Observe(api, protocol string, value float64) {
	if a.size != nil {
		a.size.With(map[string]string{
			"api":      api,
			"protocol": protocol,
		}).Observe(value)
	}
}

// APIRequestSize monitors the size of the data being received from clients.
type APIRequestSize struct {
	size *prometheus.HistogramVec
}

func NewAPIRequestSize() *APIRequestSize {
	return &APIRequestSize{
		size: metrics.RegisterHistogram("api_request_size_bytes",
			"the request sizes in bytes",
			[]string{"api", "protocol"},
			sizeBuckets),
	}
}

// Observe records the request size in bytes.
func (a APIRequestSize) Observe(api, protocol string, value float64) {
	if a.size != nil {
		a.size.With(map[string]string{
			"api":      api,
			"protocol": protocol,
		}).Observe(value)
	}
}
