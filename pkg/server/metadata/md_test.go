package metadata

import (
	"context"
	"testing"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	grpcmetadata "google.golang.org/grpc/metadata"
)

func TestGet(t *testing.T) {
	grpcCtx := grpcmetadata.NewIncomingContext(context.Background(), grpcmetadata.Pairs("id", "42", "id", "43"))
	require.Equal(t, Val("42"), Get(grpcCtx, "id"))
	require.Empty(t, Get(grpcCtx, "missing"))
	require.Empty(t, Get(context.Background(), "id"))

	requestCtx := &fasthttp.RequestCtx{}
	requestCtx.Request.Header.Add("id", "7")
	restCtx := rest.NewContext(requestCtx)
	t.Cleanup(restCtx.Close)
	require.Equal(t, Val("7"), Get(restCtx, "id"))
	require.Empty(t, Get(restCtx, "missing"))
}

func TestValConversions(t *testing.T) {
	require.Equal(t, "123", Val("123").String())
	require.Equal(t, int32(123), Val("123").Int32())
	require.Equal(t, int64(123), Val("123").Int64())
	require.Zero(t, Val("not-a-number").Int64())
}
