package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
)

const (
	// PanicInterceptorName 拦截器名称
	PanicInterceptorName = "panic"
)

func init() {
	server.AddInterceptor(PanicInterceptorName, NewPanic)
}

// Panic 奔溃拦截器
type Panic struct{}

// NewPanic panic拦截器初始化
func NewPanic() (server.ServerInterceptor, error) {
	return &Panic{}, nil
}

// Interceptor 拦截器实现
func (*Panic) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		defer func() {
			if rcv := recover(); rcv != nil {
				args := []any{
					"err", rcv,
					"req", req,
					"method", info.FullMethod,
					"protocol", info.Protocol,
					"stack", string(debug.Stack()),
				}
				logger.L().WithContext(ctx).Error("request panic", args...)
				err = status.InternalServerError()
			}
		}()
		return handler(ctx, req)
	}
}

// Name 拦截器名称
func (*Panic) Name() string {
	return PanicInterceptorName
}
