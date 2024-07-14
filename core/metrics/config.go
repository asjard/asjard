package metrics

import (
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

// Config 监控配置
type Config struct {
	Enabled     bool              `json:"enabled"`
	Collectors  utils.JSONStrings `json:"collectors"`
	PushGateway PushGatewayConfig `json:"pushGateway"`
}

type PushGatewayConfig struct {
	Endpoint string             `json:"endpoint"`
	Interval utils.JSONDuration `json:"interval"`
}

var defaultConfig = Config{
	Collectors: utils.JSONStrings{
		// "go_collector",
		// "process_collector",
		// "db_default",
		"api_request_total",
	},
}

// 获取配置
func GetConfig() Config {
	conf := defaultConfig
	config.GetWithUnmarshal(constant.ConfigMetricsPrefix, &conf)
	return conf
}
