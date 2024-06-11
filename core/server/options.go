package server

// ServerOptions 服务参数
type ServerOptions struct {
	// 服务端拦截器
	Interceptor        UnaryServerInterceptor
	HealthCheckHanlder any
}
