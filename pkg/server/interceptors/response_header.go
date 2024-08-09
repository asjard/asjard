package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
)

const (
	// HeaderResponseRequestMethod 请求方法返回头
	HeaderResponseRequestMethod = "x-request-method"
	// HeaderResponseRequestID 请求ID返回头
	HeaderResponseRequestID           = "x-request-id"
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
func NewResponseHeaderInterceptor() server.ServerInterceptor {
	return &ResponseHeader{}
}

// Interceptor .
func (ResponseHeader) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rc := ctx.(*rest.Context)
		// rc.Response.Header.Add(HeaderResponseRequestID, uuid.NewString())
		if info != nil {
			rc.Response.Header.Add(HeaderResponseRequestMethod, info.FullMethod)
		}
		return handler(ctx, req)
	}
}
