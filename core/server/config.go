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

// AddressConfig 监听地址配置
type AddressConfig struct {
	// 本地监听地址
	Listen string `json:"listen"`
	// 广播地址,主要用来垮区域通信
	Advertise string `json:"advertise"`
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	BuiltInInterceptors:    utils.JSONStrings{"trace", "ratelimiter", "metrics", "panic", "accessLog", "asynqReadEntity", "restReadEntity", "i18n"},
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
	c.Interceptors = c.BuiltInInterceptors.Merge(c.Interceptors)
	c.DefaultHandlers = c.BuiltInDefaultHandlers.Merge(c.DefaultHandlers)
	return c
}
