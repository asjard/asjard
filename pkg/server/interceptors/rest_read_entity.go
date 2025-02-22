package interceptors

import (
	"context"
	"fmt"

	"github.com/asjard/asjard/core/logger"
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
func NewReadEntityInterceptor() (server.ServerInterceptor, error) {
	return &ReadEntity{}, nil
}

// Interceptor .
func (r *ReadEntity) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rtx, ok := ctx.(*rest.Context)
		if ok {
			if err := rtx.ReadEntity(req.(proto.Message)); err != nil {
				return nil, err
			}
		} else {
			logger.L().WithContext(ctx).Error("readEntity ctx must be *rest.Context",
				"current", fmt.Sprintf("%T", ctx))
		}
		return handler(ctx, req)
	}
}
