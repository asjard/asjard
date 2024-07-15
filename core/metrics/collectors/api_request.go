package collectors

import (
	"github.com/asjard/asjard/core/metrics"
	"github.com/asjard/asjard/core/runtime"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	APIRequestLables = []string{"app", "env", "version", "service", "code", "api", "protocol"}
)

type APIRequestBase struct{}

func (APIRequestBase) labelMap(code, api, protocol string) map[string]string {
	return map[string]string{
		"app":      runtime.APP,
		"env":      runtime.Environment,
		"version":  runtime.Version,
		"service":  runtime.Name,
		"code":     code,
		"api":      api,
		"protocol": protocol,
	}
}

type APIRequestCounter struct {
	APIRequestBase
	counter *prometheus.CounterVec
}

func NewAPIRequestCounter() *APIRequestCounter {
	return &APIRequestCounter{
		counter: metrics.RegisterCounter("api_requests_total",
			"The total number of handled requests",
			APIRequestLables),
	}
}

func (a APIRequestCounter) Inc(code, api, protocol string) {
	if a.counter != nil {
		a.counter.With(a.labelMap(code, api, protocol)).Inc()
	}
}

type APIRequestDuration struct {
	APIRequestBase
	summary *prometheus.SummaryVec
}

func NewAPIRequestDuratin() *APIRequestDuration {
	return &APIRequestDuration{
		summary: metrics.RegisterSummaryVec("api_requests_duration_ms",
			"The duration of handled requests",
			APIRequestLables,
			map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			}),
	}
}

func (a APIRequestDuration) Observe(code, api, protocol string, value float64) {
	if a.summary != nil {
		a.summary.With(a.labelMap(code, api, protocol)).Observe(value)
	}
}
