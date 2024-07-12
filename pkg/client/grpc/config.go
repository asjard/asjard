package grpc

import (
	"fmt"
	"time"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

// Config grpc客户端配置
type Config struct {
	client.Config
	Options OptionsConfig `json:"options"`
}

// OptionsConfig 客户端参数配置
type OptionsConfig struct {
	Keepalive KeepaliveConfig `json:"keepalive"`
}

// KeepaliveConfig keepalive配置
type KeepaliveConfig struct {
	Time                utils.JSONDuration `json:"time"`
	Timeout             utils.JSONDuration `json:"timeout"`
	PermitWithoutStream bool               `json:"permitWithoutStream"`
}

func defaultConfig() Config {
	return Config{
		Config: client.GetConfigWithProtocol(Protocol),
		Options: OptionsConfig{
			Keepalive: KeepaliveConfig{
				Time:    utils.JSONDuration{Duration: 20 * time.Second},
				Timeout: utils.JSONDuration{Duration: time.Second},
			},
		},
	}
}

func serviceConfig(serviceName string) Config {
	conf := defaultConfig()
	config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigClientWithProtocolPrefix, Protocol),
		&conf,
		config.WithChain([]string{fmt.Sprintf(constant.ConfigClientWithSevicePrefix, Protocol, serviceName)}))
	return conf
}
