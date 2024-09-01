/*
Package constant 框架定义的一些常量信息
*/
package constant

const (
	// Framework 框架名称
	Framework = "asjard"
	// FrameworkVersion 框架版本号
	FrameworkVersion = "0.7.0"
	FrameworkGithub  = "https://github.com/asjard/asjard"
	// DefaultDelimiter 默认分隔符
	DefaultDelimiter = ","
	// ConfigDelimiter 配置分隔符
	ConfigDelimiter = "."
	// DefaultCipherName 默认加解密名称
	// DefaultCipherName = "default"
	// 全部协议
	AllProtocol = "*"
)

type TraceContextKeyType int

const (
	CurrentSpanKey TraceContextKeyType = iota
)

const (
	// ConfigServerRestPrefix rest服务配置前缀
	ConfigServerRestPrefix = "asjard.servers.rest"
	// ConfigServerGrpcPrefix grpc服务配置前缀
	ConfigServerGrpcPrefix  = "asjard.servers.grpc"
	ConfigServerPporfPrefix = "asjard.servers.pprof"

	ConfigServicePrefix = "asjard.service"

	// 服务配置前缀
	ConfigServerPrefix = Framework + ".servers"
	// 协议配置前缀
	ConfigServerWithProtocolPrefix = ConfigServerPrefix + ".%s"

	// 客户端配置前缀
	ConfigClientPrefix = Framework + ".clients"
	// 客户端协议配置前缀
	ConfigClientWithProtocolPrefix = ConfigClientPrefix + ".%s"
	// 客户端服务配置前缀
	ConfigClientWithSevicePrefix = ConfigClientWithProtocolPrefix + ".%s"

	// 拦截器配置前缀
	ConfigInterceptorPrefix = Framework + ".interceptors"
	// 服务端拦截器配置前缀
	ConfigInterceptorServerPrefix         = ConfigInterceptorPrefix + ".server"
	ConfigInterceptorServerWithNamePrefix = ConfigInterceptorServerPrefix + ".%s"
	// 客户端拦截器配置前缀
	ConfigInterceptorClientPrefix         = ConfigInterceptorPrefix + ".client"
	ConfigInterceptorClientWithNamePrefix = ConfigInterceptorClientPrefix + ".%s"

	// ConfigLoggerPrefix 日志配置前缀
	ConfigLoggerPrefix = Framework + ".logger"
	// 是否开启access_log
	ConfigLoggerAccessEnabled = ConfigLoggerPrefix + ".accessEnabled"

	// 监控配置前缀
	ConfigMetricsPrefix = Framework + ".metrics"

	ConfigLoggerBannerEnabled = "asjard.logger.banner.enabled"

	ConfigRegistryFailureThreshold    = "asjard.registry.failureThreshold"
	ConfigRegistryHealthCheck         = "asjard.registry.healthCheck"
	ConfigRegistryHealthCheckInterval = "asjard.registry.healthCheckInterval"
	ConfigRegistryLocalDiscoverPrefix = "asjard.registry.localDiscover"
	ConfigRegistryAutoRegiste         = "asjard.registry.autoRegiste"
	CofigRegistryAutoDiscove          = "asjard.registry.autoDiscove"
	ConfigRegistryDelayRegiste        = "asjard.registry.delayRegiste"
	ConfigRegistryHeartbeatInterval   = "asjard.registry.heartbeatInterval"

	ConfigInterceptorClientCircuitBreakerPrefix            = "asjard.interceptors.client.circuitBreaker"
	ConfigInterceptorClientCircuitBreakerServicePrefix     = "asjard.interceptors.client.circuitBreaker.services"
	ConfigInterceptorClientCircuitBreakerMethodPrefix      = "asjard.interceptors.client.circuitBreaker.methods"
	ConfigInterceptorClientCircuitBreakerWithServicePrefix = "asjard.interceptors.client.circuitBreaker.services.%s"
	ConfigInterceptorClientCircuitBreakerWithMethodPrefix  = "asjard.interceptors.client.circuitBreaker.methods.%s"
	ConfigInterceptorClientRest2RpcContextPrefix           = "asjard.interceptors.client.rest2RpcContext"
	ConfigInterceptorServerAccessLogPrefix                 = "asjard.interceptors.server.accessLog"

	ConfigCipherAESKey = "asjard.cipher.%s.base64Key"
	ConfigCipherAESIV  = "asjard.cipher.%s.base64Iv"
)
