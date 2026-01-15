package trace

import (
	"context"

	"github.com/asjard/asjard/pkg/server/rest"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

// MetadataCarrier implements propagation.TextMapCarrier for gRPC metadata.
// It allows the OpenTelemetry SDK to read/write trace headers from gRPC contexts.
type MetadataCarrier struct {
	md *metadata.MD
}

// HeaderCarrier implements propagation.TextMapCarrier for REST/HTTP requests.
// It wraps the custom rest.Context to access HTTP headers.
type HeaderCarrier struct {
	*rest.Context
}

// Ensure HeaderCarrier strictly follows the OpenTelemetry interface at compile time.
var _ propagation.TextMapCarrier = &HeaderCarrier{}

// NewTraceCarrier is a factory function that detects the context type (REST vs gRPC)
// and returns the appropriate carrier for trace propagation.
func NewTraceCarrier(ctx context.Context) propagation.TextMapCarrier {
	// If the context is a REST context, use the HTTP header carrier.
	if rtx, ok := ctx.(*rest.Context); ok {
		return &HeaderCarrier{rtx}
	}
	// If it's a gRPC context, extract incoming metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		return &MetadataCarrier{md: &md}
	}
	// Fallback to an empty metadata carrier if no headers are present.
	return &MetadataCarrier{md: &metadata.MD{}}
}

// --- MetadataCarrier Implementation (gRPC) ---

// Get retrieves a value from gRPC metadata for a given trace key (e.g., "traceparent").
func (c *MetadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set injects a trace key-value pair into the gRPC metadata.
func (c *MetadataCarrier) Set(key string, value string) {
	c.md.Set(key, value)
}

// Keys returns all keys currently present in the gRPC metadata.
func (c *MetadataCarrier) Keys() []string {
	out := make([]string, 0, len(*c.md))
	for key := range *c.md {
		out = append(out, key)
	}
	return out
}

// --- HeaderCarrier Implementation (REST/HTTP) ---

// Get retrieves a value from HTTP Request headers.
func (c *HeaderCarrier) Get(key string) string {
	values := c.GetHeaderParam(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set adds a trace header to the outgoing HTTP request.
func (c *HeaderCarrier) Set(key string, value string) {
	c.Request.Header.Add(key, value)
}

// Keys identifies all header keys present in the HTTP request.
func (c *HeaderCarrier) Keys() []string {
	headers := c.ReadHeaderParams()
	out := make([]string, 0, len(headers))
	for key := range headers {
		out = append(out, key)
	}
	// Note: Returns out to properly list keys found.
	return out
}
