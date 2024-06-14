package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc/metadata"
)

const (
	// 配置前缀
	configPrefix = "interceptors.client.rest2RpcContext"
)

// Rest2RpcContext rest协议的context转换为rpc的Context
// 放在拦截器的最前面
type Rest2RpcContext struct {
	cfg *rest2RpcContextConfig
}

// Rest2RpcContextConfig 拦截器配置
type rest2RpcContextConfig struct {
	AllowHeaders []string `json:"allowHeaders"`
}

func init() {
	client.AddInterceptor(NewRest2RpcContext)
}

// NewRest2RpcContext context转换初始化
func NewRest2RpcContext() client.ClientInterceptor {
	rest2RpcContext := &Rest2RpcContext{
		cfg: &rest2RpcContextConfig{},
	}
	config.GetWithUnmarshal(configPrefix,
		rest2RpcContext.cfg,
		config.WithMatchWatch(configPrefix+".*", rest2RpcContext.watch))
	return rest2RpcContext
}

// Name 拦截器名称
func (Rest2RpcContext) Name() string {
	return "rest2RpcContext"
}

// Interceptor 拦截器实现
// 来源为rest，去往rpc则把rest的请求头添加在rpc的上下文中
func (r *Rest2RpcContext) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		restCtx, ok := ctx.(*rest.Context)
		if !ok || cc.Protocol() == rest.Protocol {
			return invoker(ctx, method, req, reply, cc)
		}
		md := make(metadata.MD)
		for k, v := range restCtx.ReadHeaderParams() {
			for _, alk := range r.cfg.AllowHeaders {
				if k == alk {
					md[k] = v
					break
				}
			}
		}
		return invoker(metadata.NewIncomingContext(ctx, md), method, req, reply, cc)
	}
}

func (r *Rest2RpcContext) watch(event *config.Event) {
	var cfg rest2RpcContextConfig
	if err := config.GetWithUnmarshal(configPrefix, &cfg); err == nil {
		r.cfg = &cfg
	}
}
