package server

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

type Config struct {
	Enabled         bool              `json:"enabled"`
	Interceptors    utils.JSONStrings `json:"interceptors"`
	DefaultHandlers utils.JSONStrings `json:"defaultHandlers"`
	Addresses       map[string]string `json:"addresses"`
	CertFile        string            `json:"certFile"`
	KeyFile         string            `json:"keyFile"`
}

var DefaultConfig = Config{
	Interceptors:    utils.JSONStrings{"accessLog"},
	DefaultHandlers: utils.JSONStrings{"health"},
}

// GetConfigWithProtocol 根据协议获取配置
func GetConfigWithProtocol(protocol string) Config {
	conf := DefaultConfig
	config.GetWithUnmarshal(constant.ConfigServerPrefix, &conf)
	config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigServerWithProtocolPrefix, protocol), &conf)
	return conf
}
