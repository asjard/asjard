package loadbalance

import (
	"sync"

	"github.com/asjard/asjard/core/config"
)

// Config 负载均衡配置
type Config struct {
	*Strategy
	Services map[string]*Strategy `yaml:"services"`
	sm       sync.RWMutex
}

// Strategy 配置项
type Strategy struct {
	// 负载均衡名称
	Name string `yaml:"strategy"`
}

func loadConfig() *Config {
	return &Config{
		Strategy: &Strategy{Name: config.GetString("loadbalance.name", "")},
		Services: make(map[string]*Strategy),
	}
}

// GetStrategy 获取策略
func (c *Config) GetStrategy(serviceName string) *Strategy {
	if serviceName != "" {
		c.sm.RLock()
		serviceStrategy, ok := c.Services[serviceName]
		c.sm.RUnlock()
		if ok {
			return serviceStrategy
		}
	}
	return c.Strategy
}

// 监听配置变化
func (c *Config) watch(event *config.Event) {}
