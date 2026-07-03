package trace

import (
	"context"
	"testing"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
)

func TestMetadataCarrier(t *testing.T) {
	md := metadata.Pairs("traceparent", "one", "traceparent", "two")
	carrier := &MetadataCarrier{md: &md}
	require.Equal(t, "one", carrier.Get("traceparent"))
	require.Empty(t, carrier.Get("missing"))
	carrier.Set("baggage", "value")
	require.Equal(t, "value", carrier.Get("baggage"))
	require.ElementsMatch(t, []string{"traceparent", "baggage"}, carrier.Keys())
}

func TestNewTraceCarrier(t *testing.T) {
	grpcCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("traceparent", "grpc"))
	require.Equal(t, "grpc", NewTraceCarrier(grpcCtx).Get("traceparent"))

	empty := NewTraceCarrier(context.Background())
	empty.Set("traceparent", "new")
	require.Equal(t, "new", empty.Get("traceparent"))

	requestCtx := &fasthttp.RequestCtx{}
	requestCtx.Request.Header.Add("traceparent", "rest")
	restCtx := rest.NewContext(requestCtx)
	t.Cleanup(restCtx.Close)
	carrier := NewTraceCarrier(restCtx)
	require.IsType(t, &HeaderCarrier{}, carrier)
	require.Equal(t, "rest", carrier.Get("traceparent"))
	carrier.Set("baggage", "a")
	require.Equal(t, "a", carrier.Get("baggage"))
	require.Contains(t, carrier.Keys(), "Traceparent")
}
