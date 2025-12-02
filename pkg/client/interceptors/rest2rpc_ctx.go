package interceptors

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/client/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
)

const (
	Rest2RpcContextInterceptorName = "rest2RpcContext"
)

// Rest2RpcContext rest协议的context转换为rpc的Context
// 放在拦截器的最前面
type Rest2RpcContext struct {
	cfg rest2RpcContextConfig
	cm  sync.RWMutex
}

// Rest2RpcContextConfig 拦截器配置
type rest2RpcContextConfig struct {
	allowAllHeaders     bool
	AllowHeaders        utils.JSONStrings `json:"allowHeaders"`
	BuiltInAllowHeaders utils.JSONStrings `json:"builtInAllowHeaders"`
}

var defaultRest2RpcContextConfig = rest2RpcContextConfig{
	BuiltInAllowHeaders: utils.JSONStrings{
		"x-request-region",
		"x-request-az",
		"x-request-id",
		"x-request-instance",
		"Traceparent",
		fasthttp.HeaderXForwardedFor,
		fasthttp.HeaderAuthorization,
	},
}

func (r rest2RpcContextConfig) complete() rest2RpcContextConfig {
	r.AllowHeaders = r.BuiltInAllowHeaders.Merge(r.AllowHeaders)
	return r
}

func init() {
	client.AddInterceptor(Rest2RpcContextInterceptorName, NewRest2RpcContext, grpc.Protocol)
}

// NewRest2RpcContext context转换初始化
func NewRest2RpcContext() (client.ClientInterceptor, error) {
	rest2RpcContext := Rest2RpcContext{
		cfg: defaultRest2RpcContextConfig,
	}
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientRest2RpcContextPrefix,
		&rest2RpcContext.cfg,
		config.WithWatch(rest2RpcContext.watch)); err != nil {
		logger.Error("get interceptor config fail", "interceptor", "rest2RpcContext", "err", err)
		return nil, err
	}
	rest2RpcContext.cfg = rest2RpcContext.cfg.complete()
	return &rest2RpcContext, nil
}

// Name 拦截器名称
func (*Rest2RpcContext) Name() string {
	return Rest2RpcContextInterceptorName
}

// Interceptor 拦截器实现
// 来源为rest，去往rpc则把rest的请求头添加在rpc的上下文中
func (r *Rest2RpcContext) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		rtx, ok := ctx.(*rest.Context)
		if !ok {
			return invoker(ctx, method, req, reply, cc)
		}
		md := make(metadata.MD)
		r.cm.RLock()
		defer r.cm.RUnlock()
		for _, k := range r.cfg.AllowHeaders {
			md[k] = rtx.GetHeaderParam(k)
			v := rtx.GetUserParam(k)
			if len(v) != 0 {
				md[k] = v
			}
		}
		return invoker(metadata.NewOutgoingContext(context.Background(), md),
			method, req, reply, cc)
	}
}

func (r *Rest2RpcContext) watch(event *config.Event) {
	conf := defaultRest2RpcContextConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientRest2RpcContextPrefix, &conf); err == nil {
		r.cm.Lock()
		r.cfg = conf.complete()
		r.cm.Unlock()
	}
}
