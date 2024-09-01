package server

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

type Config struct {
	Enabled bool `json:"enabled"`
	// 自定义拦截器
	Interceptors utils.JSONStrings `json:"interceptors"`
	// 内建拦截器
	// 配置的拦截器和内建拦截器合并
	BuiltInInterceptors utils.JSONStrings `json:"builtInInterceptors"`
	// 默认处理器
	DefaultHandlers utils.JSONStrings `json:"defaultHandlers"`
	// 内建默认处理器
	BuiltInDefaultHandlers utils.JSONStrings `json:"builtInDefaultHandlers"`
	// 监听地址配置
	Addresses AddressConfig `json:"addresses"`
	// 证书文件配置,相对于ASJARD_CONF_DIR/certs的相对路径
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

type AddressConfig struct {
	Listen    string `json:"listen"`
	Advertise string `json:"advertise"`
}

var DefaultConfig = Config{
	BuiltInInterceptors:    utils.JSONStrings{"ratelimiter", "metrics", "accessLog", "restReadEntity", "restResponseHeader", "i18n", "trace"},
	BuiltInDefaultHandlers: utils.JSONStrings{"default", "health", "metrics"},
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

// 去重，添加内建配置
func (c Config) complete() Config {
	interceptors := c.BuiltInInterceptors
	for _, interceptor := range c.Interceptors {
		exist := false
		for _, inc := range interceptors {
			if inc == interceptor {
				exist = true
				break
			}
		}
		if !exist {
			interceptors = append(interceptors, interceptor)
		}
	}
	c.Interceptors = interceptors
	defaultHandlers := c.BuiltInDefaultHandlers
	for _, defaultHandler := range c.DefaultHandlers {
		exist := false
		for _, dh := range defaultHandlers {
			if defaultHandler == dh {
				exist = true
				break
			}
		}
		if !exist {
			defaultHandlers = append(defaultHandlers, defaultHandler)
		}
	}
	c.DefaultHandlers = defaultHandlers
	return c
}
