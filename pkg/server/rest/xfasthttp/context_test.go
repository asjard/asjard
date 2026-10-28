package xfasthttp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestContextParametersAndLifecycle(t *testing.T) {
	raw := &fasthttp.RequestCtx{}
	raw.Request.SetRequestURI("/users?id=1&id=2")
	raw.Request.Header.SetMethod("POST")
	raw.Request.Header.SetContentType(MIME_JSON)
	raw.Request.Header.Add("X-Test", "one")
	raw.Request.SetBodyString(`{"name":"codex"}`)
	raw.SetUserValue("user", "42")
	raw.Request.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	c := NewContext(raw)
	require.Equal(t, "1", c.Query("id"))
	require.Equal(t, "one", c.Header("X-Test"))
	require.Equal(t, "42", c.UserValue("user"))
	require.Equal(t, "10.0.0.1", c.ClientIP())
	require.JSONEq(t, `{"name":"codex"}`, string(c.JSONBody()))
	require.Equal(t, []string{"1", "2"}, c.Queries()["id"])
	require.Equal(t, []string{"42"}, c.UserValues()["user"])
	require.Contains(t, c.Headers(), "X-Test")

	require.Equal(t, context.Background(), c.Context())
	require.Equal(t, "POST", c.Method())
	require.NotEmpty(t, c.Path())
	ctx := context.WithValue(context.Background(), struct{}{}, "value")
	c.SetContext(ctx)
	require.Same(t, ctx, c.Context())
	c.Close()

	c = NewContext(raw, WithI18n(true))
	require.True(t, c.IsI18nEnabled())
	c.Close()
}
