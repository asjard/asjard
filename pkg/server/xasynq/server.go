package xasynq

import (
	"context"
	"fmt"
	"reflect"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/hibiken/asynq"
)

const (
	Protocol = "asynq"
)

// AsynqServer asynq服务定义
type AsynqServer struct {
	srv     *asynq.Server
	mux     *asynq.ServeMux
	conf    Config
	options *server.ServerOptions
	app     runtime.APP
}

var (
	_             server.Server  = &AsynqServer{}
	globalHandler GlobaltHandler = &defaultGlobalHandler{}
)

func init() {
	server.AddServer(Protocol, New)
}

// New 服务初始化
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig()
	if err := config.GetWithUnmarshal("asjard.servers.asynq", &conf); err != nil {
		return nil, err
	}
	return MustNew(conf, options)
}

// MustNew 根据配置初始化
func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	if !conf.Enabled {
		return &AsynqServer{}, nil
	}
	redisConn, err := NewRedisConn(conf.Redis)
	if err != nil {
		return nil, err
	}
	return &AsynqServer{
		conf:    conf,
		options: options,
		app:     runtime.GetAPP(),
		srv: asynq.NewServer(redisConn, asynq.Config{
			Concurrency:     conf.Options.Concurrency,
			BaseContext:     globalHandler.BaseContext(),
			RetryDelayFunc:  globalHandler.RetryDelayFunc(),
			IsFailure:       globalHandler.IsFailure(),
			ErrorHandler:    globalHandler.ErrorHandler(),
			Logger:          &asynqLogger{},
			HealthCheckFunc: globalHandler.HealthCheckFunc(),
			GroupAggregator: globalHandler.GroupAggregator(),
		}),
		mux: asynq.NewServeMux(),
	}, nil
}

// AddHandler 添加处理方法
func (s *AsynqServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invalid handler, %T must implement *asynq.Handler", handler)
	}
	return s.addHandler(h)
}

// Start 服务启动
func (s *AsynqServer) Start(startErr chan error) error {
	go func() {
		if err := s.srv.Run(s.mux); err != nil {
			startErr <- fmt.Errorf("start asynq fail %v", err)
		}
	}()
	return nil
}

// Stop 停止服务
func (s *AsynqServer) Stop() {
	s.srv.Stop()
	s.srv.Shutdown()
}

// Protocol 服务协议
func (s *AsynqServer) Protocol() string {
	return Protocol
}

// ListenAddresses 服务监听地址
func (s *AsynqServer) ListenAddresses() server.AddressConfig {
	return server.AddressConfig{}
}

// Enabled 是否已启用
func (s *AsynqServer) Enabled() bool {
	return s.conf.Enabled
}

func (s *AsynqServer) addHandler(handler Handler) error {
	desc := handler.AsynqServiceDesc()
	if desc == nil {
		return nil
	}
	ht := reflect.TypeOf(desc.HandlerType).Elem()
	st := reflect.TypeOf(handler)
	if !st.Implements(ht) {
		return fmt.Errorf("found the handler of type %v that does not satisfy %v", st, ht)
	}
	for _, method := range desc.Methods {
		if method.Pattern != "" && method.Handler != nil {
			s.addRouterHandler(method.Pattern, handler, method.Handler)
		}
	}
	return nil
}

func (s *AsynqServer) addRouterHandler(fullMethodName string, svc Handler, handler handlerFunc) {
	s.mux.HandleFunc(Pattern(fullMethodName),
		func(ctx context.Context, task *asynq.Task) error {
			if _, err := handler(&Context{Context: ctx, task: task}, svc, s.options.Interceptor); err != nil {
				return err
			}
			return nil
		})
}
