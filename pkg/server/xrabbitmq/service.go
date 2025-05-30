package xrabbitmq

import "github.com/asjard/asjard/core/server"

type HandlerFunc func(ctx *Context, srv any, interceptor server.UnaryServerInterceptor) (any, error)

type ServiceDesc struct {
	ServiceName string
	HandlerType any
	Methods     []MethodDesc
}

type MethodDesc struct {
	Queue      string
	Exchange   string
	Kind       string
	Route      string
	Consumer   string
	AutoAck    bool
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoLocal    bool
	NoWait     bool
	Internal   bool
	Table      map[string]any
	Handler    HandlerFunc
}
