package server

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

type Config struct {
	Enabled      bool              `json:"enabled"`
	Interceptors utils.JSONStrings `json:"interceptors"`
	// 默认拦截器
	defaultInterceptors utils.JSONStrings
	DefaultHandlers     utils.JSONStrings `json:"defaultHandlers"`
	// 默认处理器
	defaultHandlers utils.JSONStrings
	Addresses       map[string]string `json:"addresses"`
	CertFile        string            `json:"certFile"`
	KeyFile         string            `json:"keyFile"`
}

var DefaultConfig = Config{
	defaultInterceptors: utils.JSONStrings{"metrics", "accessLog", "restReadEntity", "restResponseHeader"},
	defaultHandlers:     utils.JSONStrings{"health", "metrics"},
}

// GetConfigWithProtocol 根据协议获取配置
func GetConfigWithProtocol(protocol string) Config {
	conf := DefaultConfig
	config.GetWithUnmarshal(constant.ConfigServerPrefix,
		&conf,
		config.WithChain([]string{fmt.Sprintf(constant.ConfigServerWithProtocolPrefix, protocol)}))
	return conf.complete()
}

// GetConfig 获取服务全局配置
func GetConfig() Config {
	conf := DefaultConfig
	config.GetWithUnmarshal(constant.ConfigServerPrefix, &conf)
	return conf.complete()
}

// 去重，添加默认配置
func (c Config) complete() Config {
	interceptors := c.defaultInterceptors
	for _, dintc := range c.Interceptors {
		exist := false
		for _, intc := range interceptors {
			if intc == dintc {
				exist = true
				break
			}
		}
		if !exist {
			interceptors = append(interceptors, dintc)
		}
	}
	c.Interceptors = interceptors
	defaultHandlers := c.defaultHandlers
	for _, dh := range c.DefaultHandlers {
		exist := false
		for _, d := range defaultHandlers {
			if dh == d {
				exist = true
				break
			}
		}
		if !exist {
			defaultHandlers = append(defaultHandlers, dh)
		}
	}
	c.DefaultHandlers = defaultHandlers
	return c
}
