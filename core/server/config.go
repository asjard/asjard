package server

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

type ServerConfig struct {
	Enabled         bool              `json:"enabled"`
	Interceptors    utils.JSONStrings `json:"interceptors"`
	DefaultHandlers utils.JSONStrings `json:"defaultHandlers"`
	Addresses       map[string]string `json:"addresses"`
	CertFile        string            `json:"certFile"`
	KeyFile         string            `json:"keyFile"`
}

var defaultConfig ServerConfig = ServerConfig{
	Interceptors:    utils.JSONStrings{"accessLog"},
	DefaultHandlers: utils.JSONStrings{"health"},
}

func GetConfigWithProtocol(protocol string) ServerConfig {
	conf := defaultConfig
	config.GetWithUnmarshal(constant.ConfigServerPrefix, &conf)
	config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigServerWithProtocolPrefix, protocol), &conf)
	return conf
}
