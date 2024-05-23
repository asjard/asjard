package rest

import "github.com/valyala/fasthttp"

type methodHandler func(srv any) fasthttp.RequestHandler

// ServiceDesc represents an RPC service's specification.
type ServiceDesc struct {
	ServiceName string
	HandlerType any
	Methods     []MethodDesc
}

// MethodDesc represents an RPC service's method specification.
type MethodDesc struct {
	MethodName string
	Method     string
	Path       string
	Handler    methodHandler
}
