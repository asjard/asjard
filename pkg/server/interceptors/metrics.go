package interceptors

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/metrics/collectors"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
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

func NewMetricsInterceptor() server.ServerInterceptor {
	return &Metrics{
		requestTotal:   collectors.NewAPIRequestCounter(),
		requestLatency: collectors.NewAPIRequestLatency(),
		requestSize:    collectors.NewAPIRequestSize(),
		responseSize:   collectors.NewAPIResponseSize(),
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
		st := status.FromError(err)
		codeStr := strconv.Itoa(int(st.Code))
		fullMethod := strings.ReplaceAll(strings.Trim(info.FullMethod, "/"), "/", ".")
		m.requestTotal.Inc(codeStr, fullMethod, info.Protocol)
		m.requestLatency.Observe(fullMethod, info.Protocol, latency.Seconds())
		if rtx, ok := ctx.(*rest.Context); ok {
			m.requestSize.Observe(fullMethod, info.Protocol, float64(computeApproximateRequestSize(rtx)))
			m.responseSize.Observe(fullMethod, info.Protocol, float64(rtx.Response.Header.ContentLength()))
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
