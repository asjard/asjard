package client

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

// Config represents the client-side configuration settings.
type Config struct {
	// Loadbalance specifies the load balancing strategy name (e.g., "roundRobin", "localityRoundRobin").
	Loadbalance string `json:"loadbalance"`
	// Interceptors defines custom user-defined interceptors for the client.
	Interceptors utils.JSONStrings `json:"interceptors"`
	// BuiltInInterceptors defines the framework-provided interceptors that run by default.
	BuiltInInterceptors utils.JSONStrings `json:"builtInInterceptors"`
	// CertFile specifies the path to the client-side TLS certificate.
	CertFile string `json:"ccertFile"`
}

// DefaultConfig provides the baseline settings for all clients if no specific configuration is found.
var DefaultConfig = Config{
	Loadbalance:         "localityRoundRobin",
	BuiltInInterceptors: utils.JSONStrings{"panic", "rest2RpcContext", "validate", "errLog", "slowLog", "cycleChainInterceptor"},
}

// GetConfigWithProtocol retrieves the configuration for a specific protocol.
// It merges the default configuration with protocol-level overrides found in the config system.
func GetConfigWithProtocol(protocol string) Config {
	conf := DefaultConfig
	config.GetWithUnmarshal(constant.ConfigClientPrefix,
		&conf,
		config.WithChain([]string{fmt.Sprintf(constant.ConfigClientWithProtocolPrefix, protocol)}))
	return conf.complete()
}

// serviceConfig retrieves the configuration for a specific service under a given protocol.
// It follows a hierarchy: Default -> Protocol Global -> Service Specific.
func serviceConfig(protocol, serviceName string) Config {
	conf := GetConfigWithProtocol(protocol)
	config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigClientWithSevicePrefix, protocol, serviceName), &conf)
	return conf.complete()
}

// complete finalizes the configuration by merging built-in and custom interceptors.
// This ensures that framework essentials (like logging and validation) are always present.
func (c Config) complete() Config {
	c.Interceptors = c.BuiltInInterceptors.Merge(c.Interceptors)
	return c
}
