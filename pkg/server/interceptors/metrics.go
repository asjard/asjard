package interceptors

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/metrics/collectors"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
)

const (
	MetricsInterceptorName = "metrics"
)

// Metrics 监控拦截器
type Metrics struct {
	requestTotal    *collectors.APIRequestCounter
	requestDuration *collectors.APIRequestDuration
}

func init() {
	// 支持所有协议
	server.AddInterceptor(NewMetricsInterceptor)
}

func NewMetricsInterceptor() server.ServerInterceptor {
	return &Metrics{
		requestTotal:    collectors.NewAPIRequestCounter(),
		requestDuration: collectors.NewAPIRequestDuratin(),
	}
}

func (Metrics) Name() string {
	return MetricsInterceptorName
}

func (m Metrics) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		now := time.Now()
		resp, err = handler(ctx, req)
		latency := time.Since(now)
		go func(latency time.Duration, fullMethod, protocol string, err error) {
			st := status.FromError(err)
			codeStr := strconv.Itoa(int(st.Code))
			fullMethod = strings.ReplaceAll(strings.Trim(info.FullMethod, "/"), "/", ".")
			m.requestTotal.Inc(codeStr,
				fullMethod, protocol)
			time.Since(now)
			m.requestDuration.Observe(fullMethod,
				protocol, float64(latency.Milliseconds()))
		}(latency, info.FullMethod, info.Protocol, err)
		return resp, err
	}
}
