package interceptors

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
	mtrace "github.com/asjard/asjard/pkg/trace"
	"github.com/asjard/asjard/utils"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	TraceInterceptorName = "trace"
	// HeaderResponseRequestID 请求ID返回头
	HeaderResponseRequestID = "x-request-id"
)

// Trace 链路追踪
type Trace struct {
	tracer     trace.Tracer
	conf       *traceConfig
	app        runtime.APP
	propagator propagation.TextMapPropagator
}

type traceConfig struct {
	Enabled bool `json:"enabled"`
	// 带协议路径的地址
	// http://127.0.0.1:4318
	// grpc://127.0.0.1:4319
	Endpoint string             `json:"endpoint"`
	Timeout  utils.JSONDuration `json:"timeout"`
	CertFile string             `json:"certFile"`
	KeyFile  string             `json:"keyFile"`
	CaFile   string             `json:"cafile"`
}

var defaultTraceConfig = traceConfig{
	Timeout: utils.JSONDuration{Duration: time.Second},
}

func init() {
	server.AddInterceptor(TraceInterceptorName, NewTraceInterceptor)
}

// NewTraceInterceptor 链路追踪拦截器初始化
func NewTraceInterceptor() (server.ServerInterceptor, error) {
	trace := &Trace{
		app: runtime.GetAPP(),
	}
	if err := trace.loadAndWatch(); err != nil {
		return nil, err
	}
	return trace, nil
}

// Name 返回拦截器名称
func (*Trace) Name() string {
	return TraceInterceptorName
}

// Interceptor 链路追踪拦截器实现
func (t *Trace) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		if t.tracer == nil || !t.conf.Enabled {
			return handler(ctx, req)
		}
		carrier := mtrace.NewTraceCarrier(ctx)
		tx, span := t.tracer.Start(t.propagator.Extract(ctx, carrier),
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(semconv.ServiceName(t.app.Instance.Name)))
		defer span.End()
		t.propagator.Inject(tx, carrier)
		if rtx, ok := ctx.(*rest.Context); ok {
			rtx.Response.Header.Add(HeaderResponseRequestID, span.SpanContext().TraceID().String())
			return handler(rtx, req)
		}
		return handler(trace.ContextWithRemoteSpanContext(tx, trace.SpanContextFromContext(tx)), req)
	}
}

func (t *Trace) loadAndWatch() error {
	if err := t.load(); err != nil {
		return err
	}
	config.AddListener("asjard.trace.*", t.watch)
	return nil
}

func (t *Trace) load() error {
	conf := defaultTraceConfig
	if err := config.GetWithUnmarshal("asjard.trace", &conf); err != nil {
		return err
	}
	t.conf = &conf
	if !conf.Enabled {
		return nil
	}

	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return err
	}
	var exporter *otlptrace.Exporter
	switch u.Scheme {
	case "http", "https":
		var options []otlptracehttp.Option
		options = append(options, otlptracehttp.WithEndpoint(u.Host))
		if u.Path != "" {
			options = append(options, otlptracehttp.WithURLPath(u.Path))
		}
		if conf.Timeout.Duration != 0 {
			options = append(options, otlptracehttp.WithTimeout(conf.Timeout.Duration))
		}
		if conf.KeyFile == "" || conf.CaFile == "" || conf.CertFile == "" {
			options = append(options, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(context.Background(), options...)
	case "grpc":
		var options []otlptracegrpc.Option
		options = append(options, otlptracegrpc.WithEndpoint(u.Host))
		if conf.Timeout.Duration != 0 {
			options = append(options, otlptracegrpc.WithTimeout(conf.Timeout.Duration))
		}
		if conf.KeyFile == "" || conf.CaFile == "" || conf.CertFile == "" {
			options = append(options, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(context.Background(), options...)
	default:
		return fmt.Errorf("trace new export unsupport schema %s", u.Scheme)
	}
	if err != nil {
		return err
	}
	t.tracer = sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter)).
		Tracer(constant.Framework,
			trace.WithInstrumentationVersion(constant.FrameworkVersion))
	t.propagator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{})
	return nil
}

func (t *Trace) watch(event *config.Event) {
	if err := t.load(); err != nil {
		logger.Error("load trace fail", "err", err)
	}
}
