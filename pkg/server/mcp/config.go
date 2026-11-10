package mcp

import (
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
)

const (
	// DefaultPath is the default HTTP path served by the MCP server.
	DefaultPath = "/mcp"
)

// Config represents the complete set of configuration options for the MCP server.
// HTTP listener settings are reused from xhttp.
type Config struct {
	rest.Config
	Path string `json:"path"`
}

func defaultConfig() Config {
	conf := rest.DefaultConfig(Protocol)
	conf.Config = server.GetConfigWithProtocol(Protocol)
	// Streamable HTTP / SSE sessions stay idle longer than regular API calls.
	conf.Options.ReadTimeout = utils.JSONDuration{}
	conf.Options.ReadHeaderTimeout = utils.JSONDuration{}
	conf.Options.IdleTimeout = utils.JSONDuration{}
	return Config{
		Config: conf,
		Path:   DefaultPath,
	}
}
