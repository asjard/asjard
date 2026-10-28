package rest

import "go.opentelemetry.io/otel/propagation"

// HTTPHeaderCarrier implements propagation.TextMapCarrier for REST/HTTP requests.
type HTTPHeaderCarrier struct {
	RestContext
}

var _ propagation.TextMapCarrier = &HTTPHeaderCarrier{}

// NewHTTPHeaderCarrier wraps a RestContext as an OpenTelemetry header carrier.
func NewHTTPHeaderCarrier(ctx RestContext) *HTTPHeaderCarrier {
	return &HTTPHeaderCarrier{RestContext: ctx}
}

func (c *HTTPHeaderCarrier) Get(key string) string {
	if v := c.UserValue(key); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return c.Header(key)
}

func (c *HTTPHeaderCarrier) Set(key string, value string) {
	c.SetUserValue(key, value)
}

func (c *HTTPHeaderCarrier) Keys() []string {
	headers := c.Headers()
	out := make([]string, 0, len(headers))
	for key := range headers {
		out = append(out, key)
	}
	return out
}
