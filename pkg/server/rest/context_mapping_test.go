package rest

import (
	"context"
	"testing"
	"time"

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
	require.Equal(t, []string{"1", "2"}, c.GetQueryParam("id"))
	require.Equal(t, []string{"one"}, c.GetHeaderParam("X-Test"))
	require.Equal(t, []string{"42"}, c.GetUserParam("user"))
	require.Equal(t, "10.0.0.1", c.ClientIP())
	require.JSONEq(t, `{"name":"codex"}`, string(c.JSONBodyParams()))
	require.Equal(t, []string{"1", "2"}, c.ReadQueryParams()["id"])
	require.Equal(t, []string{"42"}, c.ReadPathParams()["user"])
	require.Contains(t, c.ReadHeaderParams(), "X-Test")

	ctx := context.WithValue(context.Background(), struct{}{}, "value")
	c.SetContext(ctx)
	require.Same(t, ctx, c.Context())
	c.Close()
}

func TestMapForm(t *testing.T) {
	type form struct {
		Name    string
		Count   int
		Enabled bool
		Ratio   float64
		When    time.Time `json:"When" time_format:"2006-01-02"`
		Tags    []string
	}
	var got form
	err := mapForm(&got, map[string][]string{
		"Name": {"test"}, "Count": {"3"}, "Enabled": {"true"}, "Ratio": {"1.5"},
		"When": {"2026-06-21"}, "Tags": {"one", "two"},
	})
	require.NoError(t, err)
	require.Equal(t, "test", got.Name)
	require.Equal(t, 3, got.Count)
	require.True(t, got.Enabled)
	require.Equal(t, 1.5, got.Ratio)
	require.Equal(t, []string{"one", "two"}, got.Tags)
	require.Equal(t, 2026, got.When.Year())

	require.Error(t, mapForm(&got, map[string][]string{"Count": {"not-an-int"}}))
	require.NoError(t, mapForm(&got, nil))
}

func TestCorsOriginValidation(t *testing.T) {
	conf := CorsConfig{AllowOrigins: []string{"https://example.com"}}
	require.True(t, corsIsOriginValid(conf, "https://example.com"))
	require.False(t, corsIsOriginValid(conf, "https://invalid.example"))
	require.True(t, corsIsOriginValid(CorsConfig{AllowOrigins: []string{"*"}, allowAllOrigins: true}, "https://anything.example"))
}
