package interceptors

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/metrics/collectors"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/status"
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
		resp, err = handler(ctx, req)
		now := time.Now()
		fullMethod := strings.ReplaceAll(strings.Trim(info.FullMethod, "/"), "/", ".")
		code, _ := status.FromError(err)
		codeStr := strconv.Itoa(int(code))
		m.requestTotal.Inc(codeStr,
			fullMethod, info.Protocol)
		m.requestDuration.Observe(codeStr,
			fullMethod, info.Protocol, float64(time.Since(now).Milliseconds()))
		return resp, err
	}
}
