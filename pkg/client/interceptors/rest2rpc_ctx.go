package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"google.golang.org/grpc/metadata"
)

const (
	Rest2RpcContextInterceptorName = "rest2RpcContext"
)

// Rest2RpcContext rest协议的context转换为rpc的Context
// 放在拦截器的最前面
type Rest2RpcContext struct {
	cfg rest2RpcContextConfig
}

// Rest2RpcContextConfig 拦截器配置
type rest2RpcContextConfig struct {
	AllowHeaders        utils.JSONStrings `json:"allowHeaders"`
	BuiltInAllowHeaders utils.JSONStrings `json:"builtInAllowHeaders"`
}

var defaultRest2RpcContextConfig = rest2RpcContextConfig{
	BuiltInAllowHeaders: utils.JSONStrings{
		"x-request-region",
		"x-request-az",
		"x-request-id",
		"x-forward-for",
	},
}

func (r rest2RpcContextConfig) complete() rest2RpcContextConfig {
	allowHeaders := r.BuiltInAllowHeaders
	for _, allowHeader := range r.AllowHeaders {
		exist := false
		for _, ah := range r.BuiltInAllowHeaders {
			if allowHeader == ah {
				exist = true
				break
			}
		}
		if !exist {
			allowHeaders = append(allowHeaders, allowHeader)
		}
	}
	r.AllowHeaders = allowHeaders
	return r
}

func init() {
	client.AddInterceptor(Rest2RpcContextInterceptorName, NewRest2RpcContext)
}

// NewRest2RpcContext context转换初始化
func NewRest2RpcContext() client.ClientInterceptor {
	rest2RpcContext := Rest2RpcContext{
		cfg: defaultRest2RpcContextConfig,
	}
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientRest2RpcContextPrefix,
		&rest2RpcContext.cfg,
		config.WithWatch(rest2RpcContext.watch)); err != nil {
		logger.Error("get interceptor config fail", "interceptor", "rest2RpcContext", "err", err)
	}
	rest2RpcContext.cfg = rest2RpcContext.cfg.complete()
	return &rest2RpcContext
}

// Name 拦截器名称
func (Rest2RpcContext) Name() string {
	return Rest2RpcContextInterceptorName
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
	conf := defaultRest2RpcContextConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientRest2RpcContextPrefix, &conf); err == nil {
		r.cfg = conf.complete()
	}
}
