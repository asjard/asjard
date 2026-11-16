package mcp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

const initializeBody = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`

func newTestMCPHandler(disableLocalhostProtection bool) http.Handler {
	srv := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	return withMcpCORS(mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
		return srv
	}, &mcpsdk.StreamableHTTPOptions{
		DisableLocalhostProtection: disableLocalhostProtection,
		JSONResponse:               true,
		Stateless:                  true,
	}))
}

func TestMCPInitializeAllowsNonLoopbackHost(t *testing.T) {
	handler := newTestMCPHandler(true)
	httpSrv := httptest.NewServer(handler)
	t.Cleanup(httpSrv.Close)

	req, err := http.NewRequest(http.MethodPost, httpSrv.URL+"/mcp", bytes.NewBufferString(initializeBody))
	require.NoError(t, err)
	req.Host = "my-laptop.local:6010"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Origin", "vscode-file://vscode-app")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
}

func TestMCPInitializeRejectedWithoutLocalhostProtectionDisabled(t *testing.T) {
	handler := newTestMCPHandler(false)
	httpSrv := httptest.NewServer(handler)
	t.Cleanup(httpSrv.Close)

	req, err := http.NewRequest(http.MethodPost, httpSrv.URL+"/mcp", bytes.NewBufferString(initializeBody))
	require.NoError(t, err)
	req.Host = "evil.example"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestMCPOptionsPreflight(t *testing.T) {
	handler := newTestMCPHandler(true)
	req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
	req.Header.Set("Origin", "https://cursor.sh")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "https://cursor.sh", rec.Header().Get("Access-Control-Allow-Origin"))
}
