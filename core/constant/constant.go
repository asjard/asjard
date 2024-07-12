package constant

const (
	// Framework 框架名称
	Framework = "asjard"
	// FrameworkVersion 框架版本号
	FrameworkVersion = "1.0.0"
	FrameworkGithub  = "https://github.com/asjard/asjard"
	// DefaultDelimiter 默认分隔符
	DefaultDelimiter = ","
	// ConfigDelimiter 配置分隔符
	ConfigDelimiter = "."
	// ServerListenAddressName 服务监听地址
	// 代表内部地址
	ServerListenAddressName = "listen"
	// ServerAdvertiseAddressName 服务广播地址名称
	// 开放地址意味着可以在外部访问
	// 例如垮AZ，APP， Region访问可以通过此地址访问
	ServerAdvertiseAddressName = "advertise"

	// ServerProtocolKey 服务协议名称
	ServerProtocolKey = "protocol"
	// ServiceNameKey 服务名称
	ServiceNameKey = "serviceName"
	// ServiceAppKey 项目
	ServiceAppKey = "app"
	// ServiceEnvKey 环境
	ServiceEnvKey = "env"
	// ServiceRegionKey 区域
	ServiceRegionKey = "region"
	// ServiceAZKey 可用区
	ServiceAZKey = "az"
	// ServiceIDKey 服务ID
	ServiceIDKey = "serviceID"
	// ServiceVersionKey 版本
	ServiceVersionKey = "version"
	// DiscoverNameKey 服务发现者
	DiscoverNameKey = "discoverName"

	// DefaultCipherName 默认加解密名称
	DefaultCipherName = "default"
)

const (
	// ConfigServerRestPrefix rest服务配置前缀
	ConfigServerRestPrefix = "asjard.servers.rest"
	// ConfigServerGrpcPrefix grpc服务配置前缀
	ConfigServerGrpcPrefix  = "asjard.servers.grpc"
	ConfigServerPporfPrefix = "asjard.servers.pprof"

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

	// ConfigServerInterceptorPrefix 服务端拦截器前缀
	ConfigServerInterceptorPrefix = "asjard.interceptors.server"

	// ConfigLoggerPrefix 配置前缀
	ConfigLoggerPrefix        = "asjard.logger"
	ConfigLoggerAccessEnabled = "asjard.logger.accessEnabled"
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

	ConfigDatabaseMysqlPrefix = "asjard.database.mysql"

	ConfigCipherAESKey = "asjard.cipher.%s.base64Key"
	ConfigCipherAESIV  = "asjard.cipher.%s.base64Iv"

	ConfigApp           = "asjard.app"
	ConfigRegion        = "asjard.region"
	ConfigAvaliablezone = "asjard.avaliablezone"
	ConfigEnvironment   = "asjard.environment"
	ConfigWebsite       = "asjard.website"
	ConfigDesc          = "asjard.desc"
	ConfigVersion       = "asjard.instance.version"
	ConfigName          = "asjard.instance.name"
	ConfigMetadata      = "asjard.instance.metadata"
)
