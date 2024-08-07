package interceptors

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
)

const (
	// accessLog拦截器名称
	AccessLogInterceptorName = "accessLog"
)

// AccessLog access日志拦截器
type AccessLog struct {
	enabled bool
	cfg     *accessLogConfig
	m       sync.RWMutex
}

type accessLogConfig struct {
	// 配置格式: [protocol://]{fullMethod}
	// 例如grpc协议的某个方法: grpc://api.v1.hello.Hello.Call
	// 或者协议无关的某个方法: api.v1.hello.Hello.Call
	// 拦截协议的所有方法: grpc
	SkipMethods    utils.JSONStrings `json:"skipMethods"`
	skipMethodsMap map[string]struct{}
}

var defaultAccessLogConfig = accessLogConfig{
	SkipMethods: utils.JSONStrings{"grpc", "asjard.api.health.Health.Check"},
}

func init() {
	server.AddInterceptor(AccessLogInterceptorName, NewAccessLogInterceptor)
}

// NewAccessLogInterceptor .
func NewAccessLogInterceptor() server.ServerInterceptor {
	accessLog := &AccessLog{}
	accessLog.loadAndWatch()
	return accessLog
}

// Name 日志拦截器名称
func (*AccessLog) Name() string {
	return AccessLogInterceptorName
}

// Interceptor 拦截器实现
// 垮协议拦截器
func (al *AccessLog) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		if !al.enabled {
			return handler(ctx, req)
		}
		fullMethod := strings.ReplaceAll(strings.Trim(info.FullMethod, "/"), "/", ".")
		if al.skipped(info.Protocol, fullMethod) {
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

func (al *AccessLog) skipped(protocol, method string) bool {
	al.m.RLock()
	defer al.m.RUnlock()
	// 是否拦截协议
	if _, ok := al.cfg.skipMethodsMap[protocol]; ok {
		return true
	}
	// 是否拦截方法
	if _, ok := al.cfg.skipMethodsMap[method]; ok {
		return true
	}
	// 是否拦截协议方法
	if _, ok := al.cfg.skipMethodsMap[protocol+"://"+method]; ok {
		return true
	}
	return false
}

func (al *AccessLog) loadAndWatch() {
	al.load()
	config.AddListener("asjard.logger.accessEnabled", al.watch)
	config.AddPatternListener("asjard.interceptors.server.accessLog.*", al.watch)
}

func (al *AccessLog) load() {
	conf := defaultAccessLogConfig
	config.GetWithUnmarshal("asjard.interceptors.server.accessLog",
		&conf)
	conf.skipMethodsMap = make(map[string]struct{}, len(conf.SkipMethods))
	for _, skipMethod := range conf.SkipMethods {
		conf.skipMethodsMap[skipMethod] = struct{}{}
	}
	al.m.Lock()
	al.enabled = config.GetBool("asjard.logger.accessEnabled", false)
	al.cfg = &conf
	al.m.Unlock()
}

func (al *AccessLog) watch(_ *config.Event) {
	al.load()
}
