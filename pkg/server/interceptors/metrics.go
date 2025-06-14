package interceptors

import (
	"context"
	"strconv"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/metrics/collectors"
	"github.com/asjard/asjard/pkg/server/rest"
)

const (
	MetricsInterceptorName = "metrics"
)

// Metrics 监控拦截器
type Metrics struct {
	requestTotal   *collectors.APIRequestCounter
	requestLatency *collectors.APIRequestLatency
	requestSize    *collectors.APIRequestSize
	responseSize   *collectors.APIResponseSize
}

func init() {
	// 支持所有协议
	server.AddInterceptor(MetricsInterceptorName, NewMetricsInterceptor)
}

func NewMetricsInterceptor() (server.ServerInterceptor, error) {
	return &Metrics{
		requestTotal:   collectors.NewAPIRequestCounter(),
		requestLatency: collectors.NewAPIRequestLatency(),
		requestSize:    collectors.NewAPIRequestSize(),
		responseSize:   collectors.NewAPIResponseSize(),
	}, nil
}

func (Metrics) Name() string {
	return MetricsInterceptorName
}

func (m Metrics) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		logger.L(ctx).Debug("start metrics interceptor", "full_method", info.FullMethod, "protocol", info.Protocol)
		now := time.Now()
		resp, err = handler(ctx, req)
		st := status.FromError(err)
		codeStr := strconv.Itoa(int(st.Code))
		m.requestTotal.Inc(codeStr, info.FullMethod, info.Protocol)
		m.requestLatency.Observe(info.FullMethod, info.Protocol, time.Since(now).Seconds())
		if rtx, ok := ctx.(*rest.Context); ok {
			m.requestSize.Observe(info.FullMethod, info.Protocol, float64(computeApproximateRequestSize(rtx)))
			m.responseSize.Observe(info.FullMethod, info.Protocol, float64(rtx.Response.Header.ContentLength()))
		}
		return resp, err
	}
}

func computeApproximateRequestSize(r *rest.Context) int {
	s := len(r.RequestURI())
	s += len(r.Method())
	s += len(r.Request.Header.Protocol())
	s += r.Request.Header.Len()
	s += r.Request.Header.ContentLength()
	return s
}
