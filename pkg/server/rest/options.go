package rest

// Option defines a function type that modifies a Context instance.
// This is used to implement the functional options pattern for flexible configuration.
type Option func(ctx *Context)

// WithWriter allows the caller to override the default response writer.
// By passing a custom Writer, you can change how data is serialized or
// formatted (e.g., XML instead of JSON) for a specific request.
func WithWriter(wrt Writer) func(ctx *Context) {
	return func(ctx *Context) {
		if wrt != nil {
			ctx.write = wrt
		}
	}
}

// WithErrPage allows the caller to set a custom error landing page URL.
// When a request fails, the framework can use this page to redirect the user
// or render a specific UI, useful for website-oriented REST services.
func WithErrPage(errPage string) func(ctx *Context) {
	return func(ctx *Context) {
		if errPage != "" {
			ctx.errPage = errPage
		}
	}
}
