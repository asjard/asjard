package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/protobuf/validatepb"
)

const (
	ValidateInterceptorName = "validate"
)

func init() {
	server.AddInterceptor(ValidateInterceptorName, NewValidateInterceptor)
}

// Validate 参数校验
type Validate struct{}

// Name 参数考验拦截器名称
func (r *Validate) Name() string {
	return ValidateInterceptorName
}

// NewValidateInterceptor 初始化序列化参数拦截器
func NewValidateInterceptor() (server.ServerInterceptor, error) {
	return &Validate{}, nil
}

// Interceptor .
func (r *Validate) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		// 参数校验
		if v, ok := req.(validatepb.Validater); ok {
			if err := v.IsValid(info.FullMethod); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}
