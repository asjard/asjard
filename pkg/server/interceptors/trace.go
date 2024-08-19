package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/constant"
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
	TraceInterceptorName = "trace"
)

// Trace 链路追踪
type Trace struct {
	conf *mtrace.Config
	app  runtime.APP

	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

func init() {
	server.AddInterceptor(TraceInterceptorName, NewTraceInterceptor)
}

// NewTraceInterceptor 链路追踪拦截器初始化
func NewTraceInterceptor() (server.ServerInterceptor, error) {
	return &Trace{
		conf: mtrace.GetConfig(),
		app:  runtime.GetAPP(),
		tracer: otel.GetTracerProvider().Tracer(constant.Framework,
			trace.WithInstrumentationVersion(constant.FrameworkVersion)),
		propagator: otel.GetTextMapPropagator(),
	}, nil
}

// Name 返回拦截器名称
func (*Trace) Name() string {
	return TraceInterceptorName
}

// Interceptor 链路追踪拦截器实现
func (t *Trace) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		if !t.conf.Enabled {
			return handler(ctx, req)
		}
		carrier := mtrace.NewTraceCarrier(ctx)
		tx, span := t.tracer.Start(t.propagator.Extract(ctx, carrier),
			info.Protocol+":"+info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(semconv.ServiceName(t.app.Instance.Name)))
		defer span.End()
		t.propagator.Inject(tx, carrier)
		if rtx, ok := ctx.(*rest.Context); ok {
			rtx.Response.Header.Add(rest.HeaderResponseRequestID, span.SpanContext().TraceID().String())
			return handler(rtx, req)
		}
		return handler(trace.ContextWithRemoteSpanContext(tx, trace.SpanContextFromContext(tx)), req)
	}
}
