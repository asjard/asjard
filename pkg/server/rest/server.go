package rest

import (
	"github.com/asjard/asjard/core/server"
)

type methodHandler func(ctx *Context, srv any, interceptor server.UnaryServerInterceptor) (any, error)

// ServiceDesc represents an RPC service's specification.
type ServiceDesc struct {
	ServiceName string
	HandlerType any
	ErrPage     string
	Writer      Writer
	Methods     []MethodDesc
}

// MethodDesc represents an RPC service's method specification.
type MethodDesc struct {
	// 接口名称
	MethodName string
	// 接口请求方法列表
	Method string
	// 接口路径
	Path string
	// 接口处理方法
	Handler methodHandler
	// 接口描述
	Desc string
}
