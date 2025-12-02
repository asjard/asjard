package trace

import (
	"context"
	"io"
	"net/url"

	"github.com/asjard/asjard/core/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Init 链路追踪初始化
func Init() error {
	conf := GetConfig()
	if !conf.Enabled {
		return nil
	}
	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return err
	}
	// var exporter *otlptrace.Exporter
	var exporter sdktrace.SpanExporter
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
		exporter, err = stdouttrace.New(stdouttrace.WithWriter(io.Discard))
		// return fmt.Errorf("trace new export unsupport schema %s", u.Scheme)
	}
	if err != nil {
		return err
	}
	app := runtime.GetAPP()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(app.Instance.Name),
			attribute.String("app", app.App),
			attribute.String("region", app.Region),
			attribute.String("az", app.AZ),
			attribute.String("environment", app.Environment),
			attribute.String("instance", app.Instance.ID),
		)))
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagator)
	return nil
}
