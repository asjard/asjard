package registry

import (
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/utils"
)

// Config aggregates all settings for the Service Registry and Discovery lifecycle.
type Config struct {
	RegisterConfig
	DiscoverConfig
}

// RegisterConfig defines how the local service instance interacts with the registry server.
type RegisterConfig struct {
	// AutoRegiste determines if the service should automatically join the registry on startup.
	AutoRegiste bool `json:"autoRegiste"`
	// DelayRegiste allows for a "warm-up" period before the service is marked as available.
	DelayRegiste utils.JSONDuration `json:"delayRegiste"`
	// Hearbeat enables active signaling to the registry to prove the service is still alive.
	Hearbeat bool `json:"hearbeat"`
	// HeartbeatInterval defines how often the heartbeat signal is sent.
	HeartbeatInterval utils.JSONDuration `json:"heartbeatInterval"`
}

// DiscoverConfig defines how the service finds and maintains the health status of upstream dependencies.
type DiscoverConfig struct {
	// AutoDiscove determines if the client should automatically fetch the service list from the registry.
	AutoDiscove bool `json:"autoDiscove"`
	// HealthCheck enables local active probing of discovered service instances.
	HealthCheck bool `json:"healthCheck"`
	// HealthCheckInterval defines the frequency of local health probes.
	HealthCheckInterval utils.JSONDuration `json:"healthCheckInterval"`
	// FailureThreshold is the number of consecutive failed probes allowed before an instance is removed.
	FailureThreshold int `json:"failureThreshold"`
}

// defaultConfig provides stable baseline settings for production environments.
var defaultConfig = Config{
	RegisterConfig: RegisterConfig{
		AutoRegiste:       true,
		HeartbeatInterval: utils.JSONDuration{Duration: 5 * time.Second},
	},
	DiscoverConfig: DiscoverConfig{
		AutoDiscove:         true,
		HealthCheckInterval: utils.JSONDuration{Duration: 10 * time.Second},
		FailureThreshold:    1,
	},
}

// GetConfig retrieves the registry settings from the global configuration manager.
// It looks for keys under the "asjard.registry" namespace and merges them with defaults.
func GetConfig() *Config {
	conf := defaultConfig
	// Unmarshal configuration from sources (YAML, ETCD, etc.) into the struct.
	if err := config.GetWithUnmarshal("asjard.registry", &conf); err != nil {
		logger.Error("get asjard.registry fail", "err", err)
	}
	logger.Debug("get registry config", "conf", conf)
	return &conf
}
