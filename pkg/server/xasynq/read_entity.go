package xasynq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
	"google.golang.org/grpc/codes"
)

const (
	AsynqReadEntityInterceptorName = "asynqReadEntity"
)

func init() {
	// 请求参数自动解析
	// server.AddInterceptor(AsynqReadEntityInterceptorName, NewAsynqReadEntityInterceptor, Protocol)
}

// AsynqReadEntity 解析参数到请求参数中
type AsynqReadEntity struct{}

// Name .
func (r *AsynqReadEntity) Name() string {
	return AsynqReadEntityInterceptorName
}

// NewAsynqReadEntityInterceptor 初始化序列化参数拦截器
func NewAsynqReadEntityInterceptor() (server.ServerInterceptor, error) {
	return &AsynqReadEntity{}, nil
}

// Interceptor .
func (r *AsynqReadEntity) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rtx, ok := ctx.(*Context)
		if ok {
			payload := rtx.Payload()
			if err := json.Unmarshal(payload, &req); err != nil {
				logger.Error("asynq read entity fail", "payload", string(payload), "err", err)
				return nil, status.Errorf(codes.InvalidArgument, "invalid request")
			}
		} else {
			logger.Error("readEntity ctx must be *xasynq.Context", "current", fmt.Sprintf("%T", ctx))
		}
		return handler(ctx, req)
	}
}
