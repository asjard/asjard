package client

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

// Config 客户端配置
type Config struct {
	// 客户端负载均很
	Loadbalances string `json:"loadbalances"`
	// 客户端拦截器
	Interceptors utils.JSONStrings `json:"interceptors"`
	// 客户端证书
	CertFile string `json:"ccertFile"`
}

var DefaultConfig = Config{
	Loadbalances: "roundRobin",
	Interceptors: utils.JSONStrings{"rest2RpcContext", "cycleChainInterceptor", "circuitBreaker"},
}

// GetConfigWithProtocol 获取协议配置
func GetConfigWithProtocol(protocol string) Config {
	conf := DefaultConfig
	config.GetWithUnmarshal(constant.ConfigClientPrefix,
		&conf,
		config.WithChain([]string{fmt.Sprintf(constant.ConfigClientWithProtocolPrefix, protocol)}))
	return conf
}

func serviceConfig(protocol, serviceName string) Config {
	conf := GetConfigWithProtocol(protocol)
	config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigClientWithSevicePrefix, protocol, serviceName), &conf)
	return conf
}
