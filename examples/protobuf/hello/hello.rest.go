package hello

import (
	"net/http"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/valyala/fasthttp"
)

// HelloRestServer rest服务需要实现的方法
type HelloRestServer interface {
	Say(ctx *rest.Context, in *Say) (*Say, error)
}

// SayHandler .
func SayHandler(srv any) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		rest.NewContext(c).
			ReadAndWrite(func(ctx *rest.Context, in any) (any, error) {
				return srv.(HelloRestServer).Say(ctx, in.(*Say))
			}, new(Say))
	}
}

// HelloRestServiceDesc .
var HelloRestServiceDesc = rest.ServiceDesc{
	ServiceName: "api.v1.Hello",
	HandlerType: (*HelloRestServer)(nil),
	Methods: []rest.MethodDesc{
		{
			MethodName: "api.v1.hello.Hello.Say",
			Methods:    []string{http.MethodGet},
			Path:       "/",
			Handler:    SayHandler,
		},
		{
			MethodName: "Say1",
			Methods:    []string{http.MethodGet},
			Path:       "/{account_id}/",
			Handler:    SayHandler,
		},
		{
			MethodName: "Say1",
			Methods:    []string{http.MethodGet},
			Path:       "/region/{region}/user/{user_id}/",
			Handler:    SayHandler,
		},
		{
			MethodName: "Say1",
			Methods:    []string{http.MethodGet},
			Path:       "/region/{region_id}/project/{project_id}/user/{user_id}",
			Handler:    SayHandler,
		},
		{
			MethodName: "Say1",
			Methods:    []string{http.MethodPost, http.MethodPut},
			Path:       "/region/{region_id}/project/{project_id}/user/{user_id}",
			Handler:    SayHandler,
		},
	},
}
