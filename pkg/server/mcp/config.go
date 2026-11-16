package mcp

import (
	"net/http"

	"github.com/asjard/asjard/v2/core/server"
	"github.com/asjard/asjard/v2/pkg/server/rest"
	"github.com/asjard/asjard/v2/utils"
)

const (
	// DefaultPath is the default HTTP path served by the MCP server.
	DefaultPath = "/mcp"
)

// Config represents the complete set of configuration options for the MCP server.
// HTTP listener settings are reused from xhttp.
type Config struct {
	rest.Config
	Path       string           `json:"path"`
	Streamable StreamableConfig `json:"streamable"`
}

// StreamableConfig tunes the MCP Streamable HTTP transport.
type StreamableConfig struct {
	// DisableLocalhostProtection skips the SDK DNS-rebinding Host check.
	// Local clients (Cursor, VS Code) often send a non-loopback Host while the
	// connection still arrives on 127.0.0.1, which the SDK rejects with 403.
	DisableLocalhostProtection bool `json:"disableLocalhostProtection"`
	// JSONResponse returns application/json instead of SSE for POST responses.
	JSONResponse bool `json:"jsonResponse"`
	// Stateless handles each request without a persistent MCP session.
	Stateless bool `json:"stateless"`
}

func defaultConfig() Config {
	conf := rest.DefaultConfig(Protocol)
	conf.Config = server.GetConfigWithProtocol(Protocol)
	// Streamable HTTP / SSE sessions stay idle longer than regular API calls.
	conf.Options.ReadTimeout = utils.JSONDuration{}
	conf.Options.ReadHeaderTimeout = utils.JSONDuration{}
	conf.Options.IdleTimeout = utils.JSONDuration{}
	// REST CORS returns 403 for disallowed Origin. MCP clients send Origin
	// during version negotiation; a 403 is reported as "the server denied access".
	conf.Cors.Disabled = true
	conf.Cors.AllowOrigins = []string{"*"}
	conf.Cors.AllowMethods = []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodDelete, http.MethodOptions,
	}
	conf.Cors.AllowHeaders = []string{
		"Origin", "Content-Length", "Content-Type", "Accept", "Authorization",
		"MCP-Protocol-Version", "Mcp-Session-Id", "Last-Event-ID",
	}
	return Config{
		Config: conf,
		Path:   DefaultPath,
		Streamable: StreamableConfig{
			DisableLocalhostProtection: true,
		},
	}
}
