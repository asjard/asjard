package interceptors

import (
	"context"
	"strings"

	"github.com/asjard/asjard/core/metrics"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricsInterceptorName = "metrics"
)

// Metrics 监控拦截器
type Metrics struct {
	requestTotal    *prometheus.CounterVec
	requestDuration *prometheus.SummaryVec
}

func init() {
	// 支持所有协议
	server.AddInterceptor(NewMetricsInterceptor)
}

func NewMetricsInterceptor() server.ServerInterceptor {
	return &Metrics{
		requestTotal: metrics.RegisterCounter("api_request_total",
			"The total number of handled api request",
			[]string{"service", "full_method", "protocol"}),
	}
}

func (Metrics) Name() string {
	return MetricsInterceptorName
}

func (m Metrics) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		fullMethod := strings.ReplaceAll(strings.Trim(info.FullMethod, "/"), "/", ".")
		if m.requestTotal != nil {
			m.requestTotal.With(map[string]string{
				"service":     runtime.Name,
				"full_method": fullMethod,
				"protocol":    info.Protocol,
			}).Inc()
		}
		return handler(ctx, req)
	}
}
