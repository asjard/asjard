package xhttp

import (
	"time"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
)

// Config represents the complete set of configuration options for the HTTP server.
type Config struct {
	rest.Config
	Options OptionsConfig `json:"options"` // Low-level net/http server tuning options
}

// OptionsConfig contains timeout and size settings for net/http.Server.
type OptionsConfig struct {
	ReadTimeout        utils.JSONDuration `json:"readTimeout"`
	ReadHeaderTimeout  utils.JSONDuration `json:"readHeaderTimeout"`
	WriteTimeout       utils.JSONDuration `json:"writeTimeout"`
	IdleTimeout        utils.JSONDuration `json:"idleTimeout"`
	MaxHeaderBytes     int                `json:"maxHeaderBytes"`
	MaxRequestBodySize int                `json:"maxRequestBodySize"`
	DisableKeepalive   bool               `json:"disableKeepalive"`
}

// DefaultConfig initializes the HTTP server with safe, production-ready default values.
func DefaultConfig(protocol string) Config {
	return Config{
		Config: rest.DefaultConfig(protocol),
		Options: OptionsConfig{
			ReadTimeout:        utils.JSONDuration{Duration: 30 * time.Second},
			ReadHeaderTimeout:  utils.JSONDuration{Duration: 10 * time.Second},
			IdleTimeout:        utils.JSONDuration{Duration: 60 * time.Second},
			MaxRequestBodySize: 20 * 1024 * 1024,
		},
	}
}
