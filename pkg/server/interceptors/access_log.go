package interceptors

// AccessLog access日志拦截器
type AccessLog struct{}

// Name 日志拦截器名称
func (AccessLog) Name() string {
	return "accessLog"
}
