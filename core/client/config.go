package client

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

// Config 客户端配置
type Config struct {
	// 客户端负载均衡策略
	Loadbalance string `json:"loadbalance"`
	// 客户端拦截器
	Interceptors utils.JSONStrings `json:"interceptors"`
	// 内建客户端拦截器
	BuiltInInterceptors utils.JSONStrings `json:"-"`
	// 客户端证书
	CertFile string `json:"ccertFile"`
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Loadbalance:         "localityRoundRobin",
	BuiltInInterceptors: utils.JSONStrings{"rest2RpcContext", "circuitBreaker", "cycleChainInterceptor"},
}

// GetConfigWithProtocol 获取协议配置
func GetConfigWithProtocol(protocol string) Config {
	conf := DefaultConfig
	config.GetWithUnmarshal(constant.ConfigClientPrefix,
		&conf,
		config.WithChain([]string{fmt.Sprintf(constant.ConfigClientWithProtocolPrefix, protocol)}))
	return conf.complete()
}

func serviceConfig(protocol, serviceName string) Config {
	conf := GetConfigWithProtocol(protocol)
	config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigClientWithSevicePrefix, protocol, serviceName), &conf)
	return conf.complete()
}

func (c Config) complete() Config {
	c.Interceptors = c.BuiltInInterceptors.Merge(c.Interceptors)
	return c
}
