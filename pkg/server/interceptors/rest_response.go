package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
)

// RestFinalResponse rest协议最终返回
type RestFinalResponse struct{}

func init() {}

func NewRestFinalResponse() server.ServerInterceptor {
	return &RestFinalResponse{}
}

func (RestFinalResponse) Name() string {
	return "restFinalResponse"
}

// Interceptor rest协议可以支持自定义返回
func (RestFinalResponse) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		if cc, ok := ctx.(*rest.Context); ok {
			resp, err := handler(ctx, req)
			cc.Write(resp, err)
		}
		return handler(ctx, req)
	}
}
