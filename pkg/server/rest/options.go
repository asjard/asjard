package rest

// Option .
type Option func(ctx *Context)

// WithWriter 设置输出方法
func WithWriter(wrt Writer) func(ctx *Context) {
	return func(ctx *Context) {
		if wrt != nil {
			ctx.write = wrt
		}
	}
}

// WithErrPage 设置默认错误页
func WithErrPage(errPage string) func(ctx *Context) {
	return func(ctx *Context) {
		if errPage != "" {
			ctx.errPage = errPage
		}
	}
}
