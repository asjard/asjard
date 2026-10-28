package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	mtrace "github.com/asjard/asjard/core/trace"
	"github.com/asjard/asjard/pkg/server/rest"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceInterceptorName is the unique identifier for the tracing interceptor.
	TraceInterceptorName = "trace"
)

// Trace handles the lifecycle of distributed tracing spans for server-side requests.
type Trace struct {
	conf *mtrace.Config
	app  runtime.APP

	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

func init() {
	// Register the tracing interceptor globally for all protocols.
	server.AddInterceptor(TraceInterceptorName, NewTraceInterceptor)
}

// NewTraceInterceptor initializes the tracing component with the global OpenTelemetry provider.
func NewTraceInterceptor() (server.ServerInterceptor, error) {
	return &Trace{
		conf: mtrace.GetConfig(),
		app:  runtime.GetAPP(),
		// Create a tracer identified by the framework name and version.
		tracer: otel.GetTracerProvider().Tracer(constant.Framework,
			trace.WithInstrumentationVersion(constant.FrameworkVersion)),
		// Use the global propagator to handle carrier extraction/injection (e.g., W3C TraceContext).
		propagator: otel.GetTextMapPropagator(),
	}, nil
}

// Name returns the interceptor name.
func (*Trace) Name() string {
	return TraceInterceptorName
}

// Interceptor implements the tracing logic for every unary request.
func (t *Trace) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		logger.L(ctx).Debug("start server interceptor", "interceptor", t.Name(), "full_method", info.FullMethod, "protocol", info.Protocol)
		// Skip if tracing is disabled in configuration.
		if !t.conf.Enabled {
			return handler(ctx, req)
		}

		// Prepare the carrier (headers/metadata) and extract existing trace context from the caller.
		carrier := mtrace.NewTraceCarrier(ctx)
		rtx, isREST := ctx.(rest.RestContext)
		if isREST {
			carrier = rest.NewHTTPHeaderCarrier(rtx)
		}

		// Start a new Span.
		// Operation name format: {protocol}://{fullMethod} (e.g., grpc://api.User/Get).
		tx, span := t.tracer.Start(t.propagator.Extract(ctx, carrier),
			info.Protocol+"://"+info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(semconv.ServiceName(t.app.Instance.Name), semconv.ServiceNamespace(t.app.Instance.Group)))
		defer span.End()

		// Inject the new span context back into the carrier for downstream propagation.
		t.propagator.Inject(tx, carrier)

		// Special handling for REST/HTTP protocol to attach TraceID to the response.
		if isREST {
			rtx.SetContext(trace.ContextWithSpanContext(rtx.Context(), trace.SpanContextFromContext(tx)))
			return handler(ctx, req)
		}

		// Pass the enriched context containing the span to the next handler.
		return handler(trace.ContextWithRemoteSpanContext(tx, trace.SpanContextFromContext(tx)), req)
	}
}
