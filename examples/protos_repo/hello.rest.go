package account

import (
	"net/http"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/valyala/fasthttp"
)

// AccountRestServer rest服务需要实现的方法
type AccountRestServer interface {
	Say(ctx *rest.Context, in *Empty) (*SayResp, error)
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
	},
}
