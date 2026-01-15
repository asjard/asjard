package grpc

import (
	"fmt"
	"time"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

// Config represents the complete configuration for a gRPC client.
// It embeds the common client.Config and adds gRPC-specific options.
type Config struct {
	client.Config
	// Options contains specific dial and connection parameters.
	Options OptionsConfig `json:"options"`
}

// OptionsConfig holds the gRPC-specific connection tuning parameters.
type OptionsConfig struct {
	// Keepalive settings to ensure connection health.
	Keepalive KeepaliveConfig `json:"keepalive"`
}

// KeepaliveConfig defines the parameters for client-side keepalive pings.
// These settings help maintain connections through load balancers and firewalls.
type KeepaliveConfig struct {
	// Time is the interval between keepalive pings.
	Time utils.JSONDuration `json:"time"`
	// Timeout is the time the client waits for a response before closing the connection.
	Timeout utils.JSONDuration `json:"timeout"`
	// PermitWithoutStream allows pings even when there are no active RPC calls.
	PermitWithoutStream bool `json:"permitWithoutStream"`
}

// defaultConfig returns the baseline settings for any gRPC client created by the framework.
func defaultConfig() Config {
	return Config{
		// Inherit general client protocol settings (like timeout, retry, etc.)
		Config: client.GetConfigWithProtocol(Protocol),
		Options: OptionsConfig{
			Keepalive: KeepaliveConfig{
				// Default to sending a ping every 20s with a 3s response window.
				Time:    utils.JSONDuration{Duration: 20 * time.Second},
				Timeout: utils.JSONDuration{Duration: 3 * time.Second},
			},
		},
	}
}

// serviceConfig merges global gRPC defaults with service-specific overrides.
// It uses a configuration chain: Default -> Global gRPC -> Specific Service.
func serviceConfig(serviceName string) Config {
	conf := defaultConfig()
	// Unmarshal configuration from the provider (e.g., YAML/ETCD).
	// Priority 1: asjard.clients.grpc.services.{serviceName}
	// Priority 2: asjard.clients.grpc
	config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigClientWithProtocolPrefix, Protocol),
		&conf,
		config.WithChain([]string{fmt.Sprintf(constant.ConfigClientWithSevicePrefix, Protocol, serviceName)}))
	return conf
}
