package rest

// Option .
type Option func(ctx *Context)

// WithWriter .
func WithWriter(wrt Writer) func(ctx *Context) {
	return func(ctx *Context) {
		if wrt != nil {
			ctx.write = wrt
		}
	}
}

// WithErrPage .
func WithErrPage(errPage string) func(ctx *Context) {
	return func(ctx *Context) {
		if errPage != "" {
			ctx.errPage = errPage
		}
	}
}
