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
	// MetricsInterceptorName is the unique identifier for this interceptor.
	MetricsInterceptorName = "metrics"
)

// Metrics handles the telemetry collection for every incoming request.
type Metrics struct {
	requestTotal   *collectors.APIRequestCounter // Counts total requests by code, method, and protocol.
	requestLatency *collectors.APIRequestLatency // Tracks the duration of request execution.
	requestSize    *collectors.APIRequestSize    // Tracks the size of incoming request payloads.
	responseSize   *collectors.APIResponseSize   // Tracks the size of outgoing response payloads.
}

func init() {
	// Register the metrics interceptor to support all protocols (gRPC and REST).
	server.AddInterceptor(MetricsInterceptorName, NewMetricsInterceptor)
}

// NewMetricsInterceptor initializes the Prometheus collectors.
func NewMetricsInterceptor() (server.ServerInterceptor, error) {
	return &Metrics{
		requestTotal:   collectors.NewAPIRequestCounter(),
		requestLatency: collectors.NewAPIRequestLatency(),
		requestSize:    collectors.NewAPIRequestSize(),
		responseSize:   collectors.NewAPIResponseSize(),
	}, nil
}

// Name returns the interceptor's unique name.
func (Metrics) Name() string {
	return MetricsInterceptorName
}

// Interceptor returns the middleware function that records telemetry data.
func (m Metrics) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		logger.L(ctx).Debug("start server interceptor", "interceptor", m.Name(), "full_method", info.FullMethod, "protocol", info.Protocol)

		start := time.Now()

		// 1. Execute the business logic handler.
		resp, err = handler(ctx, req)

		// 2. Extract the response status code (converts gRPC status to string).
		st := status.FromError(err)
		codeStr := strconv.Itoa(int(st.Code))

		// 3. Record Golden Signals:
		// - Increment the counter for total requests.
		m.requestTotal.Inc(codeStr, info.FullMethod, info.Protocol)

		// - Record execution latency in seconds.
		m.requestLatency.Observe(info.FullMethod, info.Protocol, float64(time.Now().Sub(start))/float64(time.Second))

		// 4. Protocol-specific metrics (Size tracking for REST).
		if rtx, ok := ctx.(*rest.Context); ok {
			// Observe approximate size of the HTTP request.
			m.requestSize.Observe(info.FullMethod, info.Protocol, float64(computeApproximateRequestSize(rtx)))
			// Observe the exact content length of the HTTP response.
			m.responseSize.Observe(info.FullMethod, info.Protocol, float64(rtx.Response.Header.ContentLength()))
		}

		return resp, err
	}
}

// computeApproximateRequestSize calculates the total byte size of an HTTP request
// including URI, Method, Protocol version, and Headers.
func computeApproximateRequestSize(r *rest.Context) int {
	s := len(r.RequestURI())
	s += len(r.Method())
	s += len(r.Request.Header.Protocol())
	s += r.Request.Header.Len()
	s += r.Request.Header.ContentLength()
	return s
}
