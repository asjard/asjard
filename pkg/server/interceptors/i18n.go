package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
)

const (
	I18nInterceptorName = "i18n"
)

// I18n i18n拦截器
type I18n struct {
}

func init() {
	// 支持所有协议
	server.AddInterceptor(NewMetricsInterceptor, rest.Protocol)
}

// I18n拦截器初始化
func NewI18nInterceptor() server.ServerInterceptor {
	return &I18n{}
}

func (I18n) Name() string {
	return I18nInterceptorName
}

func (m I18n) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err == nil {
			return resp, err
		}
		_, ok := ctx.(*rest.Context)
		if !ok {
			return resp, err
		}
		return resp, err
	}
}
