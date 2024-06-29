package constant

const (
	// Framework 框架名称
	Framework = "asjard"
	// FrameworkVersion 框架版本号
	FrameworkVersion = "1.0.0"
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
	// ConfigLoggerPrefix 配置前缀
	ConfigLoggerPrefix        = "asjard.logger"
	ConfigLoggerAccessEnabled = "asjard.logger.accessEnabled"
	ConfigLoggerBannerEnabled = "asjard.logger.banner.enabled"

	// ConfigBalanceWithProtocol 负载协议相关配置
	ConfigBalanceWithProtocol = "asjard.clients.%s.loadbalances"
	// ConfigBalanceWithProtocolAndService 负载均衡服相关配置
	ConfigBalanceWithProtocolAndService = "asjard.clients.%s.%s.loadbalances"
	// ConfigBalance 负载通用配置
	ConfigBalance = "asjard.clients.loadbalances"

	// ConfigClientInterceptorWithProtocol 客户端协议相关拦截器配置
	ConfigClientInterceptorWithProtocol           = "asjard.clients.%s.interceptors"
	ConfigClientInterceptorWithProtocolAndService = "asjard.clients.%s.%s.interceptors"
	// ConfigClientInterceptor 客户端通用拦截器配置
	ConfigClientInterceptor                                        = "asjard.clients.interceptors"
	ConfigClientCertFileWithProtocol                               = "asjard.clients.%s.certFile"
	ConfigClientCertFileWithProtocolAndService                     = "asjard.clients.%s.%s.certFile"
	ConfigClientGrpcOptionsKeepaliveTimeWithService                = "asjard.clients.grpc.%s.options.keepalive.Time"
	ConfigClientGrpcOptionsKeepaliveTime                           = "clients.grpc.options.keepalive.Time"
	ConfigClientGrpcOptionsKeepaliveTimeoutWithService             = "asjard.clients.grpc.%s.options.keepalive.Timeout"
	ConfigClientGrpcOptionsKeepaliveTimeout                        = "clients.grpc.options.keepalive.Timeout"
	ConfigClientGrpcOptionsKeepalivePermitWithoutStreamWithService = "asjard.clients.grpc.%s.options.keepalive.PermitWithoutStream"
	ConfigClientGrpcOptionsKeepalivePermitWithoutStream            = "clients.grpc.options.keepalive.PermitWithoutStream"

	ConfigServerDefaultHandlersWithProtocol = "asjard.servers.%s.defaultHandlers"
	ConfigServerDefaultHandlers             = "asjard.servers.defaultHandlers"
	ConfigServerInterceptorWithProtocol     = "asjard.servers.%s.interceptors"
	ConfigServerInterceptor                 = "asjard.servers.interceptors"
	ConfigServerAddress                     = "asjard.servers.%s.addresses"
	ConfigServerCertfile                    = "asjard.servers.%s.certFile"
	ConfigServerKeyFile                     = "asjard.servers.%s.keyFile"
	ConfigServerEnabled                     = "asjard.servers.%s.enabled"
	ConfigServerGrpcOptionsMaxConnectIdle   = "asjard.servers.grpc.options.MaxConnectionIdle"
	ConfigServerGrpcOptionsTime             = "asjard.servers.grpc.options.Time"
	ConfigServerGrpcOptionsTimeout          = "asjard.servers.grpc.options.Timeout"
	ConfigServerRestDocErrPage              = "asjard.servers.rest.doc.errPage"

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
	ConfigVersion       = "asjard.instance.version"
	ConfigName          = "asjard.instance.name"
	ConfigMetadata      = "asjard.instance.metadata"
)
