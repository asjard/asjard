package rest

// RouterHandler .
type RouterHandler func(ctx *Context) (any, error)

// Router .
type Router struct {
	// 路由名称
	Name string
	// 路由
	Path string
	// 请求方法, GET, POST, PUT, DELETE
	Method string
	// 路由handler
	Handler RouterHandler
}
