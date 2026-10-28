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

		// Execute the business logic handler.
		resp, err = handler(ctx, req)

		// Extract the response status code (converts gRPC status to string).
		st := status.FromError(err)
		codeStr := strconv.Itoa(int(st.Code))

		// Record Golden Signals:
		// - Increment the counter for total requests.
		m.requestTotal.Inc(codeStr, info.FullMethod, info.Protocol)

		// - Record execution latency in seconds.
		// m.requestLatency.Observe(info.FullMethod, info.Protocol, float64(time.Now().Sub(start))/float64(time.Second))
		m.requestLatency.Observe(info.FullMethod, info.Protocol, time.Since(start).Seconds())

		if mtx, ok := ctx.(rest.RestContext); ok {
			m.requestSize.Observe(info.FullMethod, info.Protocol, float64(mtx.RequestSize()))
			m.responseSize.Observe(info.FullMethod, info.Protocol, float64(mtx.ResponseSize()))
		}

		return resp, err
	}
}
