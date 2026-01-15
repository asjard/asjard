package server

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

// Config defines the settings for a server instance.
type Config struct {
	// Enabled indicates if this server should be started.
	Enabled bool `json:"enabled"`

	// Interceptors are custom middleware specified by the user in config files.
	Interceptors utils.JSONStrings `json:"interceptors"`

	// BuiltInInterceptors are standard framework middleware (e.g., tracing, logging).
	// These are merged with custom Interceptors during initialization.
	BuiltInInterceptors utils.JSONStrings `json:"builtInInterceptors"`

	// DefaultHandlers are custom API handlers specified by the user.
	DefaultHandlers utils.JSONStrings `json:"defaultHandlers"`

	// BuiltInDefaultHandlers are system-provided handlers (e.g., /health, /metrics).
	BuiltInDefaultHandlers utils.JSONStrings `json:"builtInDefaultHandlers"`

	// Addresses specifies where the server listens and how it announces itself.
	Addresses AddressConfig `json:"addresses"`

	// CertFile is the relative path to the TLS certificate (relative to ASJARD_CONF_DIR/certs).
	CertFile string `json:"certFile"`
	// KeyFile is the relative path to the TLS private key.
	KeyFile string `json:"keyFile"`
}

// AddressConfig manages network binding and service discovery announcement.
type AddressConfig struct {
	// Listen is the local network address the server binds to (e.g., "0.0.0.0:8080").
	Listen string `json:"listen"`

	// Advertise is the address broadcasted to the service registry.
	// This is useful for cross-region communication or NAT/Docker environments.
	Advertise string `json:"advertise"`
}

// DefaultConfig provides a standard baseline for all servers in the framework.
var DefaultConfig = Config{
	// Standard interceptor stack order: panic recovery -> i18n -> trace -> etc.
	BuiltInInterceptors: utils.JSONStrings{"panic", "i18n", "trace", "ratelimiter", "metrics", "accessLog", "restReadEntity"},
	// Standard diagnostic and monitoring endpoints.
	BuiltInDefaultHandlers: utils.JSONStrings{"default", "health", "metrics"},
}

// GetConfigWithProtocol retrieves server settings for a specific protocol (e.g., "rest", "grpc").
// It follows a configuration chain: Global Server Config -> Protocol-Specific Config.
func GetConfigWithProtocol(protocol string) Config {
	conf := DefaultConfig
	config.GetWithUnmarshal(constant.ConfigServerPrefix,
		&conf,
		config.WithChain([]string{fmt.Sprintf(constant.ConfigServerWithProtocolPrefix, protocol)}))
	return conf.complete()
}

// GetConfig retrieves the global server settings without protocol-specific overrides.
func GetConfig() Config {
	conf := DefaultConfig
	config.GetWithUnmarshal(constant.ConfigServerPrefix, &conf)
	return conf.complete()
}

// complete handles the merging logic to ensure built-in and custom configurations coexist correctly.
// It deduplicates lists and ensures that essential framework features are included.
func (c Config) complete() Config {
	c.Interceptors = c.BuiltInInterceptors.Merge(c.Interceptors)
	c.DefaultHandlers = c.BuiltInDefaultHandlers.Merge(c.DefaultHandlers)
	return c
}
