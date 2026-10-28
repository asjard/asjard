package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
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
}
