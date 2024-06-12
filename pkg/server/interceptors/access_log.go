package interceptors

import (
	"context"
	"strings"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
)

// AccessLog access日志拦截器
type AccessLog struct {
	enabled bool
	cfg     *accessLogConfig
}

type accessLogConfig struct {
	SkipMethods    []string `json:"skipMethods"`
	skipMethodsMap map[string]struct{}
}

func init() {
	server.AddInterceptor(NewAccessLogInterceptor)
}

// NewAccessLogInterceptor .
func NewAccessLogInterceptor() server.ServerInterceptor {
	accessLog := &AccessLog{
		enabled: config.GetBool("logger.accessEnabled", false),
		cfg: &accessLogConfig{
			skipMethodsMap: make(map[string]struct{}),
		},
	}
	if err := config.GetWithUnmarshal("interceptors.server.accessLog", accessLog.cfg); err == nil {
		for _, method := range accessLog.cfg.SkipMethods {
			accessLog.cfg.skipMethodsMap[method] = struct{}{}
		}
	} else {
		logger.Error("get interceptors.server.accessLog fail", "err", err.Error())
	}
	return accessLog
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
		fullMethod := strings.ReplaceAll(strings.Trim(info.FullMethod, "/"), "/", ".")
		// 是否拦截方法
		if _, ok := al.cfg.skipMethodsMap[fullMethod]; ok {
			return handler(ctx, req)
		}
		// 是否拦截协议方法
		if _, ok := al.cfg.skipMethodsMap[info.Protocol+":"+fullMethod]; ok {
			return handler(ctx, req)
		}
		// 是否拦截协议
		if _, ok := al.cfg.skipMethodsMap[info.Protocol]; ok {
			return handler(ctx, req)
		}
		now := time.Now()
		var fields []any
		fields = append(fields, []any{"protocol", info.Protocol}...)
		fields = append(fields, []any{"full_method", fullMethod}...)
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
		if err != nil {
			logger.Error("access log", fields...)
		} else {
			logger.Info("access log", fields...)
		}
		return resp, err
	}
}
