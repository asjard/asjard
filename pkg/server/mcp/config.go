package mcp

import (
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest/xhttp"
	"github.com/asjard/asjard/utils"
)

const (
	// DefaultPath is the default HTTP path served by the MCP server.
	DefaultPath = "/mcp"
)

// Config represents the complete set of configuration options for the MCP server.
// HTTP listener settings are reused from xhttp.
type Config struct {
	xhttp.Config
	Path string `json:"path"`
}

func defaultConfig() Config {
	httpConf := xhttp.DefaultConfig(xhttp.Protocol)
	httpConf.Config.Config = server.GetConfigWithProtocol(Protocol)
	// Streamable HTTP / SSE sessions stay idle longer than regular API calls.
	httpConf.Options.ReadTimeout = utils.JSONDuration{}
	httpConf.Options.ReadHeaderTimeout = utils.JSONDuration{}
	httpConf.Options.IdleTimeout = utils.JSONDuration{}
	return Config{
		Config: httpConf,
		Path:   DefaultPath,
	}
}
