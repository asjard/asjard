/*
 * Package trace 链路追踪，添加描述
 */
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

// Init initializes the distributed tracing system.
// It sets up the exporter (where to send data), the TracerProvider (how to process data),
// and the global propagator (how to share trace IDs across services).
func Init() error {
	conf := GetConfig()
	// Exit early if tracing is disabled in configuration.
	if !conf.Enabled {
		return nil
	}

	// Parse the collector endpoint to determine the protocol (HTTP vs gRPC).
	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return err
	}

	var exporter sdktrace.SpanExporter
	switch u.Scheme {
	case "http", "https":
		// Configure OTLP/HTTP exporter.
		var options []otlptracehttp.Option
		options = append(options, otlptracehttp.WithEndpoint(u.Host))
		if u.Path != "" {
			options = append(options, otlptracehttp.WithURLPath(u.Path))
		}
		if conf.Timeout.Duration != 0 {
			options = append(options, otlptracehttp.WithTimeout(conf.Timeout.Duration))
		}
		// Fallback to insecure if TLS certificates are not provided.
		if conf.KeyFile == "" || conf.CaFile == "" || conf.CertFile == "" {
			options = append(options, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(context.Background(), options...)

	case "grpc":
		// Configure OTLP/gRPC exporter.
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
		// Fallback exporter that discards data if the scheme is unrecognized.
		exporter, err = stdouttrace.New(stdouttrace.WithWriter(io.Discard))
	}

	if err != nil {
		return err
	}

	// Fetch current application metadata from the runtime package.
	app := runtime.GetAPP()

	// Initialize the TracerProvider.
	// 1. WithSampler(AlwaysSample): Capture 100% of traces.
	// 2. WithBatcher: Buffer spans and send them in batches for better performance.
	// 3. WithResource: Attach global attributes (App, Region, Env) to every span.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(app.Instance.Name),
			attribute.String("app", app.App),
			attribute.String("region", app.Region),
			attribute.String("az", app.AZ),
			attribute.String("environment", app.Environment),
			attribute.String("instance", app.Instance.ID),
		)),
	)

	// Configure the global Propagator to support both W3C TraceContext and Baggage.
	// This ensures Trace IDs are correctly passed between different microservices.
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	// Set as the global OpenTelemetry defaults so all libraries can use them.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagator)

	return nil
}
