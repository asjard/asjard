package interceptors

import (
	"context"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
)

// AccessLog access日志拦截器
type AccessLog struct {
	enabled bool
}

func init() {
	server.AddInterceptor(NewAccessLogInterceptor)
}

// NewAccessLogInterceptor .
func NewAccessLogInterceptor() server.ServerInterceptor {
	return &AccessLog{
		enabled: config.GetBool("logger.accessEnabled", false),
	}
}

// Name 日志拦截器名称
func (AccessLog) Name() string {
	return "accessLog"
}

// Interceptor 拦截器实现
// 垮协议拦截器
func (al AccessLog) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		if !al.enabled {
			return handler(ctx, req)
		}
		now := time.Now()
		var fields []any
		fields = append(fields, []any{"protocol", info.Protocol}...)
		fields = append(fields, []any{"full_method", info.FullMethod}...)
		switch info.Protocol {
		case rest.Protocol:
			rc := ctx.(*rest.Context)
			fields = append(fields, []any{"header", rc.ReadHeaderParams()}...)
			fields = append(fields, []any{"method", string(rc.Method())}...)
			fields = append(fields, []any{"path", string(rc.Path())}...)
		}
		resp, err = handler(ctx, req)
		fields = append(fields, []any{"cost", time.Since(now).String()}...)
		fields = append(fields, []any{"req", req}...)
		fields = append(fields, []any{"success", err == nil}...)
		fields = append(fields, []any{"err", err}...)
		logger.Info("access log", fields...)
		return resp, err
	}
}
