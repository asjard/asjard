package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/pkg/client/grpc"
	"github.com/asjard/asjard/pkg/protobuf/validatepb"
)

const (
	ValidateInterceptorName = "validate"
)

// Validate 参数校验
type Validate struct{}

func init() {
	client.AddInterceptor(ValidateInterceptorName, NewValidateInterceptor, grpc.Protocol)
}

// Name 参数考验拦截器名称
func (r *Validate) Name() string {
	return ValidateInterceptorName
}

// NewValidateInterceptor 初始化序列化参数拦截器
func NewValidateInterceptor() (client.ClientInterceptor, error) {
	return &Validate{}, nil
}

func (r *Validate) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		if v, ok := req.(validatepb.Validater); ok {
			if err := v.IsValid(method); err != nil {
				return err
			}
		}
		return invoker(ctx, method, req, reply, cc)
	}
}
