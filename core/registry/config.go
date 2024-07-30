package registry

import (
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/utils"
)

// Config 注册发现中心配置
type Config struct {
	RegisterConfig
	DiscoverConfig
}

// 注册配置
type RegisterConfig struct {
	AutoRegiste       bool               `json:"autoRegiste"`
	DelayRegiste      utils.JSONDuration `json:"delayRegiste"`
	Hearbeat          bool               `json:"hearbeat"`
	HeartbeatInterval utils.JSONDuration `json:"heartbeatInterval"`
}

// 发现配置
type DiscoverConfig struct {
	AutoDiscove         bool               `json:"autoDiscove"`
	HealthCheck         bool               `json:"healthCheck"`
	HealthCheckInterval utils.JSONDuration `json:"healthCheckInterval"`
	FailureThreshold    int                `json:"FailureThreshold"`
}

var defaultConfig = Config{
	RegisterConfig: RegisterConfig{
		AutoRegiste:       true,
		HeartbeatInterval: utils.JSONDuration{Duration: 5 * time.Second},
	},
	DiscoverConfig: DiscoverConfig{
		AutoDiscove:         true,
		HealthCheckInterval: utils.JSONDuration{Duration: 10 * time.Second},
	},
}

// 获取配置
func GetConfig() *Config {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.registry", &conf); err != nil {
		logger.Error("get asjard.registry fail", "err", err)
	}
	logger.Debug("get registry config", "conf", conf)
	return &conf
}
