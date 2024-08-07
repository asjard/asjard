package metrics

import (
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

const (
	AllCollectors = "*"
)

// Config 监控配置
type Config struct {
	Enabled bool `json:"enabled"`
	// 是否是所有指标
	allCollectors     bool
	Collectors        utils.JSONStrings `json:"collectors"`
	BuiltInCollectors utils.JSONStrings `json:"BuiltInCollectors"`
	PushGateway       PushGatewayConfig `json:"pushGateway"`
}

type PushGatewayConfig struct {
	Endpoint string             `json:"endpoint"`
	Interval utils.JSONDuration `json:"interval"`
}

var defaultConfig = Config{
	BuiltInCollectors: utils.JSONStrings{
		"go_collector",
		"process_collector",
		"db_default",
		"api_requests_total",
		"api_requests_latency_seconds",
		"api_request_size_bytes",
		"api_response_size_bytes",
	},
	PushGateway: PushGatewayConfig{
		Interval: utils.JSONDuration{Duration: 5 * time.Second},
	},
}

// 获取配置
func GetConfig() Config {
	conf := defaultConfig
	config.GetWithUnmarshal(constant.ConfigMetricsPrefix, &conf)
	return conf.complete()
}

func (c Config) complete() Config {
	collectors := c.BuiltInCollectors
	for _, collector := range c.Collectors {
		if collector == AllCollectors {
			c.allCollectors = true
			break
		}
		exist := false
		for _, ct := range c.BuiltInCollectors {
			if ct == collector {
				exist = true
				break
			}
		}
		if !exist {
			collectors = append(collectors, collector)
		}
	}
	c.Collectors = collectors
	return c
}
