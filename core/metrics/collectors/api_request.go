package collectors

import (
	"github.com/asjard/asjard/core/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

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

type APIRequestDuration struct {
	summary *prometheus.SummaryVec
}

// Ref: http://dimacs.rutgers.edu/~graham/pubs/papers/bquant-icde.pdf
func NewAPIRequestDuratin() *APIRequestDuration {
	return &APIRequestDuration{
		summary: metrics.RegisterSummaryVec("api_requests_latency_ms",
			"The duration of handled requests",
			[]string{"api", "protocol"},
			map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			}),
	}
}

func (a APIRequestDuration) Observe(api, protocol string, value float64) {
	if a.summary != nil {
		a.summary.With(map[string]string{
			"api":      api,
			"protocol": protocol,
		}).Observe(value)
	}
}
