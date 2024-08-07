package collectors

import (
	"github.com/asjard/asjard/core/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// APIRequestCounter 请求数量
type APIRequestCounter struct {
	counter *prometheus.CounterVec
}

func NewAPIRequestCounter() *APIRequestCounter {
	return &APIRequestCounter{
		counter: metrics.RegisterCounter("api_requests_total",
			"The total number of handled requests",
			[]string{"code", "api", "protocol"}),
	}
}

func (a APIRequestCounter) Inc(code, api, protocol string) {
	if a.counter != nil {
		a.counter.With(map[string]string{
			"code":     code,
			"api":      api,
			"protocol": protocol,
		}).Inc()
	}
}

// APIRequestLatenccy 请求耗时
type APIRequestLatency struct {
	latency *prometheus.HistogramVec
}

// Ref: http://dimacs.rutgers.edu/~graham/pubs/papers/bquant-icde.pdf
func NewAPIRequestLatency() *APIRequestLatency {
	return &APIRequestLatency{
		latency: metrics.RegisterHistogram("api_requests_latency_seconds",
			"The duration of handled requests",
			[]string{"api", "protocol"},
			prometheus.DefBuckets),
	}
}

func (a APIRequestLatency) Observe(api, protocol string, value float64) {
	if a.latency != nil {
		a.latency.With(map[string]string{
			"api":      api,
			"protocol": protocol,
		}).Observe(value)
	}
}

const (
	_           = iota // ignore first value by assigning to blank identifier
	bKB float64 = 1 << (10 * iota)
	bMB
)

var sizeBuckets = []float64{1.0 * bKB, 2.0 * bKB, 5.0 * bKB, 10.0 * bKB, 100 * bKB, 500 * bKB, 1.0 * bMB, 2.5 * bMB, 5.0 * bMB, 10.0 * bMB}

// APIResponseSize 返回大小
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

func (a APIResponseSize) Observe(api, protocol string, value float64) {
	if a.size != nil {
		a.size.With(map[string]string{
			"api":      api,
			"protocol": protocol,
		}).Observe(value)
	}
}

// APIRequestSize 请求大小
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

func (a APIRequestSize) Observe(api, protocol string, value float64) {
	if a.size != nil {
		a.size.With(map[string]string{
			"api":      api,
			"protocol": protocol,
		}).Observe(value)
	}
}
