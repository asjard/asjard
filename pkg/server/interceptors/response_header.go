package interceptors

import (
	"context"
	"fmt"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
)

const (
	// HeaderResponseRequestMethod 请求方法返回头
	HeaderResponseRequestMethod       = "x-request-method"
	RestResponseHeaderInterceptorName = "restResponseHeader"
)

func init() {
	// 统一添加返回头
	server.AddInterceptor(RestResponseHeaderInterceptorName, NewResponseHeaderInterceptor, rest.Protocol)
}

// ResponseHeader 添加返回头
type ResponseHeader struct{}

// Name .
func (ResponseHeader) Name() string {
	return RestResponseHeaderInterceptorName
}

// NewResponseHeaderInterceptor 初始化返回请求头拦截器
func NewResponseHeaderInterceptor() (server.ServerInterceptor, error) {
	return &ResponseHeader{}, nil
}

// Interceptor .
func (ResponseHeader) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rtx, ok := ctx.(*rest.Context)
		if ok {
			if info != nil {
				rtx.Response.Header.Add(HeaderResponseRequestMethod, info.FullMethod)
			}
		} else {
			logger.Error("readEntity ctx must be *rest.Context", "current", fmt.Sprintf("%T", ctx))
		}
		return handler(ctx, req)
	}
}
