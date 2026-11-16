package metadata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asjard/asjard/v2/pkg/server/rest"
	"github.com/stretchr/testify/require"
	grpcmetadata "google.golang.org/grpc/metadata"
)

func TestGet(t *testing.T) {
	grpcCtx := grpcmetadata.NewIncomingContext(context.Background(), grpcmetadata.Pairs("id", "42", "id", "43"))
	require.Equal(t, Val("42"), Get(grpcCtx, "id"))
	require.Empty(t, Get(grpcCtx, "missing"))
	require.Empty(t, Get(context.Background(), "id"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Add("id", "7")
	restCtx := rest.NewContext(req, httptest.NewRecorder())
	require.Equal(t, Val("7"), Get(restCtx, "id"))
	require.Empty(t, Get(restCtx, "missing"))
}

func TestValConversions(t *testing.T) {
	require.Equal(t, "123", Val("123").String())
	require.Equal(t, int32(123), Val("123").Int32())
	require.Equal(t, int64(123), Val("123").Int64())
	require.Zero(t, Val("not-a-number").Int64())
}
