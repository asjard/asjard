package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/protobuf/proto"
)

const (
	RestReadEntityInterceptorName = "restReadEntity"
)

func init() {
	// 请求参数自动解析
	server.AddInterceptor(RestReadEntityInterceptorName, NewReadEntityInterceptor, rest.Protocol)
}

// ReadEntity 解析参数到请求参数中
type ReadEntity struct{}

// Name .
func (r *ReadEntity) Name() string {
	return RestReadEntityInterceptorName
}

// NewReadEntityInterceptor 初始化序列化参数拦截器
func NewReadEntityInterceptor() server.ServerInterceptor {
	return &ReadEntity{}
}

// Interceptor .
func (r *ReadEntity) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rc := ctx.(*rest.Context)
		if err := rc.ReadEntity(req.(proto.Message)); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}
