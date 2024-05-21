package rest

import "context"

// Router .
type Router struct {
	// 路由名称
	Name string
	// 路由
	Path string
	// 请求方法, GET, POST, PUT, DELETE
	Method string
	// 路由handler
	Handler func(ctx context.Context, in any) (any, error)
}

// Group .
type Group struct {
}
