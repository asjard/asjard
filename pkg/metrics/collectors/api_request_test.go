package collectors

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestCollectorsRecordValues(t *testing.T) {
	counter := APIRequestCounter{counter: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_requests_total"}, []string{"code", "api", "protocol"})}
	counter.Inc("200", "/test", "rest")

	latency := APIRequestLatency{latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "test_latency", Buckets: []float64{1}}, []string{"api", "protocol"})}
	latency.Observe("/test", "grpc", 0.5)
	requestSize := APIRequestSize{size: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "test_request_size", Buckets: sizeBuckets}, []string{"api", "protocol"})}
	requestSize.Observe("/test", "rest", 100)
	responseSize := APIResponseSize{size: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "test_response_size", Buckets: sizeBuckets}, []string{"api", "protocol"})}
	responseSize.Observe("/test", "rest", 200)
	require.NotEmpty(t, sizeBuckets)
}

func TestNilCollectorsAreSafe(t *testing.T) {
	require.NotPanics(t, func() {
		(APIRequestCounter{}).Inc("200", "/", "rest")
		(APIRequestLatency{}).Observe("/", "rest", 1)
		(APIRequestSize{}).Observe("/", "rest", 1)
		(APIResponseSize{}).Observe("/", "rest", 1)
	})
}
