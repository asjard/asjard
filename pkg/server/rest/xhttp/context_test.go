package xhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestContextParametersAndLifecycle(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users/42?id=1&id=2", bytes.NewBufferString(`{"name":"codex"}`))
	req.Header.Set(headerContentType, MIME_JSON)
	req.Header.Add("X-Test", "one")
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	req.SetPathValue("user", "42")
	req.Pattern = "/users/{user}"
	rec := httptest.NewRecorder()

	c := NewContext(NewRequestCtx(rec, req))
	require.False(t, c.IsI18nEnabled())
	require.Equal(t, http.MethodPost, c.Method())
	require.Equal(t, "/users/42", c.Path())
	require.Equal(t, "1", c.Query("id"))
	require.Equal(t, "one", c.Header("X-Test"))
	require.Equal(t, "42", c.UserValue("user"))
	require.Equal(t, "10.0.0.1", c.ClientIP())
	require.JSONEq(t, `{"name":"codex"}`, string(c.JSONBody()))
	require.Equal(t, []string{"1", "2"}, c.Queries()["id"])
	require.Equal(t, []string{"42"}, c.UserValues()["user"])
	require.Contains(t, c.Headers(), "X-Test")
	require.Greater(t, c.RequestSize(), 0)

	require.Equal(t, context.Background(), c.Context())
	ctx := context.WithValue(context.Background(), struct{}{}, "value")
	c.SetContext(ctx)
	require.Same(t, ctx, c.Context())

	cancelReq := httptest.NewRequest(http.MethodGet, "/", nil)
	reqCtx, cancel := context.WithCancel(context.Background())
	cancelReq = cancelReq.WithContext(reqCtx)
	c = NewContext(NewRequestCtx(httptest.NewRecorder(), cancelReq))
	cancel()
	select {
	case <-c.Done():
		require.ErrorIs(t, c.Err(), context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("xhttp context did not observe request cancel")
	}
	c.Close()

	c = NewContext(NewRequestCtx(rec, req), WithI18n(true))
	require.True(t, c.IsI18nEnabled())
	c.Write([]byte("ok"))
	require.Equal(t, 2, c.ResponseSize())
	c.Close()
}

func TestRoutePattern(t *testing.T) {
	require.Equal(t, "GET /users/{id}", routePattern(http.MethodGet, "/users/{id}"))
	require.Equal(t, "/users/{id}", routePattern("", "/users/{id}"))
	require.Equal(t, []string{"id", "postId"}, pathParamNames("GET /users/{id}/posts/{postId}"))
	require.False(t, isMethodNotAllowed("GET /hello/{name}", http.MethodHead))
	require.True(t, isMethodNotAllowed("GET /hello/{name}", http.MethodDelete))
}
