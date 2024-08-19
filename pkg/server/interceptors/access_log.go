package interceptors

import (
	"context"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/client/grpc"
	"github.com/asjard/asjard/pkg/protobuf/healthpb"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
)

const (
	// accessLog拦截器名称
	AccessLogInterceptorName = "accessLog"
)

// AccessLog access日志拦截器
type AccessLog struct {
	cfg *accessLogConfig
	m   sync.RWMutex
}

type accessLogConfig struct {
	Enabled bool `json:"enabled"`
	// 配置格式: [protocol://]{fullMethod}
	// 例如grpc协议的某个方法: grpc://api.v1.hello.Hello.Call
	// 或者协议无关的某个方法: api.v1.hello.Hello.Call
	// 拦截协议的所有方法: grpc
	SkipMethods    utils.JSONStrings `json:"skipMethods"`
	skipMethodsMap map[string]struct{}
}

var defaultAccessLogConfig = accessLogConfig{
	Enabled:     true,
	SkipMethods: utils.JSONStrings{grpc.Protocol, healthpb.Health_Check_FullMethodName},
}

func init() {
	server.AddInterceptor(AccessLogInterceptorName, NewAccessLogInterceptor)
}

// NewAccessLogInterceptor .
func NewAccessLogInterceptor() (server.ServerInterceptor, error) {
	accessLog := &AccessLog{}
	if err := accessLog.loadAndWatch(); err != nil {
		return nil, err
	}
	return accessLog, nil
}

// Name 日志拦截器名称
func (*AccessLog) Name() string {
	return AccessLogInterceptorName
}

// Interceptor 拦截器实现
// 垮协议拦截器
func (al *AccessLog) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		if !al.cfg.Enabled {
			return handler(ctx, req)
		}
		if al.skipped(info.Protocol, info.FullMethod) {
			return handler(ctx, req)
		}
		now := time.Now()
		var fields []any
		fields = append(fields, []any{"protocol", info.Protocol}...)
		fields = append(fields, []any{"full_method", info.FullMethod}...)
		switch info.Protocol {
		case rest.Protocol:
			if rc, ok := ctx.(*rest.Context); ok {
				fields = append(fields, []any{"header", rc.ReadHeaderParams()}...)
				fields = append(fields, []any{"method", string(rc.Method())}...)
				fields = append(fields, []any{"path", string(rc.Path())}...)
			}
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

func (al *AccessLog) loadAndWatch() error {
	if err := al.load(); err != nil {
		return err
	}
	config.AddPatternListener("asjard.logger.accessLog.*", al.watch)
	return nil
}

func (al *AccessLog) load() error {
	conf := defaultAccessLogConfig
	if err := config.GetWithUnmarshal("asjard.logger.accessLog",
		&conf); err != nil {
		return err
	}
	conf.skipMethodsMap = make(map[string]struct{}, len(conf.SkipMethods))
	for _, skipMethod := range conf.SkipMethods {
		conf.skipMethodsMap[skipMethod] = struct{}{}
	}
	al.m.Lock()
	al.cfg = &conf
	al.m.Unlock()
	return nil
}

func (al *AccessLog) watch(_ *config.Event) {
	al.load()
}
