package xasynq

import (
	"strings"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
)

type handlerFunc func(ctx *Context, srv any, interceptor server.UnaryServerInterceptor) (any, error)

// ServiceDesc represents an RPC service's specification.
type ServiceDesc struct {
	ServiceName string
	HandlerType any
	Methods     []MethodDesc
}

// MethodDesc represents an asynq service's method specification.
type MethodDesc struct {
	// 接口名称
	Pattern string
	// 接口处理方法
	Handler handlerFunc
}

// Pattern fulleMethodName 转为asynq的pattern
func Pattern(fulleMethodName string) string {
	pattern := strings.TrimPrefix(strings.ReplaceAll(fulleMethodName, "/", ":"), ":")
	return runtime.GetAPP().ResourceKey(Protocol, pattern, runtime.WithoutService(true), runtime.WithDelimiter(":"))
}
