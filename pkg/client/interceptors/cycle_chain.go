package interceptors

import (
	"context"
	"strings"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/client/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const (
	// HeaderRequestChain 调用链 {protocol}://{app}/{serviceName}/{method}
	HeaderRequestChain = "x-request-chain"
	// HeaderRequestDest 请求目的地
	HeaderRequestDest = "x-request-dest"
	// HeaderRequestApp 请求应用
	HeaderRequestApp          = "x-request-app"
	CycleChainInterceptorName = "cycleChainInterceptor"
)

// CycleChainInterceptor 循环调用拦截器
// 依赖loadbalance在上下文注入x-request-dest和x-request-app
type CycleChainInterceptor struct {
}

func init() {
	client.AddInterceptor(CycleChainInterceptorName, NewCycleChainInterceptor, grpc.Protocol)
}

// CycleChainInterceptor 初始化来源拦截器
func NewCycleChainInterceptor() (client.ClientInterceptor, error) {
	return &CycleChainInterceptor{}, nil
}

// Name 拦截器名称
func (CycleChainInterceptor) Name() string {
	return CycleChainInterceptorName
}

// Interceptor 拦截器
// 上下文中添加当前服务
// 如果出现循环服务则拦截
// 当前只支持目的地为grpc的 rest -> grpc -> grpc
func (s CycleChainInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		if _, ok := cc.(*grpc.ClientConn); !ok {
			return invoker(ctx, method, req, reply, cc)
		}
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md, ok = metadata.FromIncomingContext(ctx)
			if !ok {
				md = metadata.New(map[string]string{})
			}
		}
		currentRequestMethod := "grpc://" + strings.ReplaceAll(strings.Trim(method, "/"), "/", ".")
		if requestChains, ok := md[HeaderRequestChain]; ok {
			for _, requestMethod := range requestChains {
				if requestMethod == currentRequestMethod {
					requestChains = append(requestChains, currentRequestMethod)
					return status.Errorf(codes.Canceled, "cycle call, chains: %s", strings.Join(requestChains, " -> "))
				}
			}
		}
		md[HeaderRequestChain] = append(md[HeaderRequestChain], currentRequestMethod)
		return invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc)

	}
}
