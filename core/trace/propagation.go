package trace

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

// MetadataCarrier implements propagation.TextMapCarrier for gRPC metadata.
// It allows the OpenTelemetry SDK to read/write trace headers from gRPC contexts.
type MetadataCarrier struct {
	md *metadata.MD
}

var _ propagation.TextMapCarrier = &MetadataCarrier{}

// NewTraceCarrier is a factory function that detects gRPC metadata on the context
// and returns the appropriate carrier for trace propagation.
func NewTraceCarrier(ctx context.Context) propagation.TextMapCarrier {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		return &MetadataCarrier{md: &md}
	}
	return &MetadataCarrier{md: &metadata.MD{}}
}

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
