package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/pkg/server/rest/xfasthttp"
	"github.com/asjard/asjard/pkg/server/rest/xhttp"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestHTTPHeaderCarrier(t *testing.T) {
	requestCtx := &fasthttp.RequestCtx{}
	requestCtx.Request.Header.Add("traceparent", "rest")
	restCtx := xfasthttp.NewContext(requestCtx)
	t.Cleanup(restCtx.Close)
	carrier := rest.NewHTTPHeaderCarrier(restCtx)
	require.Equal(t, "rest", carrier.Get("traceparent"))
	carrier.Set("baggage", "a")
	require.Equal(t, "a", carrier.Get("baggage"))
	require.Contains(t, carrier.Keys(), "Traceparent")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("traceparent", "http")
	httpCtx := xhttp.NewContext(xhttp.NewRequestCtx(httptest.NewRecorder(), req))
	t.Cleanup(httpCtx.Close)
	httpCarrier := rest.NewHTTPHeaderCarrier(httpCtx)
	require.Equal(t, "http", httpCarrier.Get("traceparent"))
}
