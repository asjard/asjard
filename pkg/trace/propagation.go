package trace

import (
	"context"

	"github.com/asjard/asjard/pkg/server/rest"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

type MetadataCarrier struct {
	md *metadata.MD
}

type HeaderCarrier struct {
	*rest.Context
}

var _ propagation.TextMapCarrier = &HeaderCarrier{}

func NewTraceCarrier(ctx context.Context) propagation.TextMapCarrier {
	if rtx, ok := ctx.(*rest.Context); ok {
		return &HeaderCarrier{rtx}
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		return &MetadataCarrier{md: &md}
	}
	return &MetadataCarrier{md: &metadata.MD{}}
}

func (c *MetadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
func (c *MetadataCarrier) Set(key string, value string) {
	c.md.Set(key, value)
}

func (c *MetadataCarrier) Keys() []string {
	out := make([]string, 0, len(*c.md))
	for key := range *c.md {
		out = append(out, key)
	}
	return out
}

func (c *HeaderCarrier) Get(key string) string {
	values := c.GetHeaderParam(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c *HeaderCarrier) Set(key string, value string) {
	c.Request.Header.Add(key, value)
}

func (c *HeaderCarrier) Keys() []string {
	headers := c.ReadHeaderParams()
	out := make([]string, 0, len(headers))
	for key := range headers {
		out = append(out, key)
	}
	return []string{}
}
