package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/runtime"
	mtrace "github.com/asjard/asjard/core/trace"
	"github.com/asjard/asjard/pkg/client/grpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	TraceInterceptorName = "trace"
)

type Trace struct {
	conf *mtrace.Config
	app  runtime.APP

	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

func init() {
	client.AddInterceptor(TraceInterceptorName, NewTrace, grpc.Protocol)
}

func NewTrace() (client.ClientInterceptor, error) {
	return &Trace{
		conf: mtrace.GetConfig(),
		app:  runtime.GetAPP(),
		tracer: otel.GetTracerProvider().Tracer(constant.Framework,
			trace.WithInstrumentationVersion(constant.FrameworkVersion)),
		propagator: otel.GetTextMapPropagator(),
	}, nil
}

func (Trace) Name() string {
	return TraceInterceptorName
}

// Interceptor grpc客户端拦截器
func (t *Trace) Interceptor() client.UnaryClientInterceptor {
	tracer := otel.GetTracerProvider().Tracer(constant.Framework,
		trace.WithInstrumentationVersion(constant.FrameworkVersion))
	propagator := otel.GetTextMapPropagator()
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		if !t.conf.Enabled {
			return invoker(ctx, method, req, reply, cc)
		}
		carrier := mtrace.NewTraceCarrier(ctx)
		ctx, span := tracer.Start(propagator.Extract(ctx, carrier),
			method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(semconv.ServiceName(t.app.Instance.Name)))
		defer span.End()
		propagator.Inject(ctx, carrier)
		return invoker(ctx, method, req, reply, cc)
	}
}
