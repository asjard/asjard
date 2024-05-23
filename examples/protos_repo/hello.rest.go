package account

import (
	"net/http"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/valyala/fasthttp"
)

// AccountRestServer rest服务需要实现的方法
type AccountRestServer interface {
	Say(ctx *rest.Context, in *Empty) (*SayResp, error)
	rest.Handler
}

// AccountSayHandler .
func AccountSayHandler(srv any) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		in := new(Empty)
		ctx := rest.NewContext(c)
		if err := ctx.ReadEntity(in); err != nil {
			ctx.Write(nil, err)
			return
		}
		ctx.Write(srv.(AccountRestServer).Say(ctx, in))
	}
}

// HelloRestServiceDesc .
var HelloRestServiceDesc = rest.ServiceDesc{
	ServiceName: "api.account.Account",
	HandlerType: (*AccountRestServer)(nil),
	Methods: []rest.MethodDesc{
		{
			MethodName: "Say",
			Method:     http.MethodGet,
			Path:       "/",
			Handler:    AccountSayHandler,
		},
		{
			MethodName: "Say1",
			Method:     http.MethodGet,
			Path:       "/{account_id}/",
			Handler:    AccountSayHandler,
		},
		{
			MethodName: "Say1",
			Method:     http.MethodGet,
			Path:       "/region/{region}/user/{user_id}/",
			Handler:    AccountSayHandler,
		},
		{
			MethodName: "Say1",
			Method:     http.MethodGet,
			Path:       "/region/{region_id}/project/{project_id}/user/{user_id}",
			Handler:    AccountSayHandler,
		},
		{
			MethodName: "Say1",
			Method:     http.MethodPost,
			Path:       "/region/{region_id}/project/{project_id}/user/{user_id}",
			Handler:    AccountSayHandler,
		},
	},
}
