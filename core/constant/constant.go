/*
Package constant defines shared framework constants and global configuration keys.
Centralizing these ensures consistency across the core, listeners, and external plugins.
*/
package constant

import (
	"sync/atomic"
)

const (
	// Framework identity and versioning.
	Framework        = "asjard"
	FrameworkVersion = "1.1.0"
	FrameworkGithub  = "https://github.com/asjard/asjard"

	// Delimiters for data processing.
	// DefaultDelimiter is used when splitting strings into slices (e.g., CSV).
	DefaultDelimiter = ","
	// ConfigDelimiter defines the hierarchy separator for configuration keys.
	ConfigDelimiter = "."

	// AllProtocol represents a wildcard for all communication protocols (gRPC, REST, etc.).
	AllProtocol = "*"
)

// TraceContextKeyType defines a custom type for context keys to avoid collisions.
type TraceContextKeyType int

const (
	// CurrentSpanKey is used to store/retrieve the current tracing span from a context.
	CurrentSpanKey TraceContextKeyType = iota
)

var (
	// Global atomic variables representing the deployment environment.
	// These are stored as atomic.Value to allow safe concurrent updates during bootstrap.
	APP         atomic.Value // Application Name
	Region      atomic.Value // Physical Region (e.g., us-east-1)
	AZ          atomic.Value // Availability Zone (e.g., az-1)
	Env         atomic.Value // Deployment Environment (e.g., prod, dev, staging)
	ServiceName atomic.Value // Specific Microservice Name
)

const (
	// Server Configuration Namespaces
	ConfigServerRestPrefix  = "asjard.servers.rest"
	ConfigServerGrpcPrefix  = "asjard.servers.grpc"
	ConfigServerPporfPrefix = "asjard.servers.pprof"
	ConfigServicePrefix     = "asjard.service"

	// Dynamic Server/Protocol Key Generators
	ConfigServerPrefix             = Framework + ".servers"
	ConfigServerWithProtocolPrefix = ConfigServerPrefix + ".%s"

	// Client Configuration Namespaces
	ConfigClientPrefix             = Framework + ".clients"
	ConfigClientWithProtocolPrefix = ConfigClientPrefix + ".%s"
	ConfigClientWithSevicePrefix   = ConfigClientWithProtocolPrefix + ".%s"

	// Interceptor (Middleware) Configuration
	ConfigInterceptorPrefix               = Framework + ".interceptors"
	ConfigInterceptorServerPrefix         = ConfigInterceptorPrefix + ".server"
	ConfigInterceptorServerWithNamePrefix = ConfigInterceptorServerPrefix + ".%s"
	ConfigInterceptorClientPrefix         = ConfigInterceptorPrefix + ".client"
	ConfigInterceptorClientWithNamePrefix = ConfigInterceptorClientPrefix + ".%s"

	// Logging and Banner settings
	ConfigLoggerPrefix        = Framework + ".logger"
	ConfigLoggerAccessEnabled = ConfigLoggerPrefix + ".accessEnabled"
	ConfigLoggerBannerDisable = "asjard.logger.banner.disable"

	// Metrics and Monitoring
	ConfigMetricsPrefix = Framework + ".metrics"

	// Service Registry and Discovery parameters
	ConfigRegistryFailureThreshold    = "asjard.registry.failureThreshold"
	ConfigRegistryHealthCheck         = "asjard.registry.healthCheck"
	ConfigRegistryHealthCheckInterval = "asjard.registry.healthCheckInterval"
	ConfigRegistryLocalDiscoverPrefix = "asjard.registry.localDiscover"
	ConfigRegistryAutoRegiste         = "asjard.registry.autoRegiste"
	CofigRegistryAutoDiscove          = "asjard.registry.autoDiscove"
	ConfigRegistryDelayRegiste        = "asjard.registry.delayRegiste"
	ConfigRegistryHeartbeatInterval   = "asjard.registry.heartbeatInterval"

	// Resilience (Circuit Breaker) and Observability Interceptors
	ConfigInterceptorClientCircuitBreakerPrefix            = "asjard.interceptors.client.circuitBreaker"
	ConfigInterceptorClientCircuitBreakerServicePrefix     = "asjard.interceptors.client.circuitBreaker.services"
	ConfigInterceptorClientCircuitBreakerMethodPrefix      = "asjard.interceptors.client.circuitBreaker.methods"
	ConfigInterceptorClientCircuitBreakerWithServicePrefix = "asjard.interceptors.client.circuitBreaker.services.%s"
	ConfigInterceptorClientCircuitBreakerWithMethodPrefix  = "asjard.interceptors.client.circuitBreaker.methods.%s"
	ConfigInterceptorClientRest2RpcContextPrefix           = "asjard.interceptors.client.rest2RpcContext"
	ConfigInterceptorClientSlowLogPrefix                   = "asjard.interceptors.client.slowLog"
	ConfigInterceptorClientErrLogPrefix                    = "asjard.interceptors.client.errLog"
	ConfigInterceptorServerAccessLogPrefix                 = "asjard.interceptors.server.accessLog"

	// Security/Cryptography Keys
	// %s represents the cipher instance name (e.g., 'default').
	ConfigCipherAESKey = "asjard.cipher.%s.base64Key"
	ConfigCipherAESIV  = "asjard.cipher.%s.base64Iv"
)
