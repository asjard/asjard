package interceptors

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"google.golang.org/protobuf/proto"
)

const (
	RestReadEntityInterceptorName = "restReadEntity"
)

func init() {
	// 请求参数自动解析
	server.AddInterceptor(RestReadEntityInterceptorName, NewReadEntityInterceptor, rest.Protocol)
}

// ReadEntity 解析参数到请求参数中
type ReadEntity struct {
	cm   sync.RWMutex
	conf ReadEntityConfig
}

// 配置
type ReadEntityConfig struct {
	SkipMethods   utils.JSONStrings `json:"skipMethods"`
	skipMethodMap map[string]struct{}
}

// Name .
func (r *ReadEntity) Name() string {
	return RestReadEntityInterceptorName
}

// NewReadEntityInterceptor 初始化序列化参数拦截器
func NewReadEntityInterceptor() (server.ServerInterceptor, error) {
	readEntity := &ReadEntity{}
	if err := readEntity.loadAndWatch(); err != nil {
		return nil, err
	}
	return readEntity, nil
}

// Interceptor .
func (r *ReadEntity) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rtx, ok := ctx.(*rest.Context)
		if ok && !r.isSkipped(info.FullMethod) {
			if err := rtx.ReadEntity(req.(proto.Message)); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

func (r *ReadEntity) isSkipped(method string) bool {
	r.cm.RLock()
	defer r.cm.RUnlock()
	_, ok := r.conf.skipMethodMap[method]
	return ok
}

func (r *ReadEntity) loadAndWatch() error {
	if err := r.load(); err != nil {
		return err
	}
	config.AddPrefixListener("asjard.interceptors.server.restReadEntity", r.watch)
	return nil
}

func (r *ReadEntity) load() error {
	conf := ReadEntityConfig{
		skipMethodMap: map[string]struct{}{},
	}
	if err := config.GetWithUnmarshal("asjard.interceptors.server.restReadEntity", &conf); err != nil {
		return err
	}
	for _, item := range conf.SkipMethods {
		conf.skipMethodMap[item] = struct{}{}
	}
	r.cm.Lock()
	r.conf = conf
	r.cm.Unlock()
	return nil
}

func (r *ReadEntity) watch(event *config.Event) {
	if err := r.load(); err != nil {
		logger.Error("read entity load config fail", "err", err)
	}
}
