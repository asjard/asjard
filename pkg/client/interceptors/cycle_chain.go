package interceptors

import (
	"context"
	"strings"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// HeaderRequestChain 调用链 {protocol}://{app}/{serviceName}/{method}
	HeaderRequestChain = "x-request-chain"
)

// CycleChainInterceptor 循环调用拦截器
type CycleChainInterceptor struct {
	currentServiceName string
}

func init() {
	client.AddInterceptor(NewCycleChainInterceptor)
}

// CycleChainInterceptor 初始化来源拦截器
func NewCycleChainInterceptor() client.ClientInterceptor {
	return &CycleChainInterceptor{
		currentServiceName: runtime.APP + "/" + runtime.Name,
	}
}

// Name 拦截器名称
func (CycleChainInterceptor) Name() string {
	return "cycleChainInterceptor"
}

// Interceptor 拦截器
// 上下文中添加当前服务
// 如果出现循环服务则拦截
// 当前只支持目的地为grpc的 rest -> grpc -> grpc
func (s CycleChainInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc grpc.ClientConnInterface, invoker client.UnaryInvoker) error {
		if _, ok := cc.(*grpc.ClientConn); !ok {
			return invoker(ctx, method, req, reply, cc)
		}
		md := make(map[string][]string)
		if rctx, ok := ctx.(*rest.Context); ok {
			// 来源为rest
			md[HeaderRequestChain] = rctx.GetHeaderParam(HeaderRequestChain)
		} else {
			md, _ = metadata.FromIncomingContext(ctx)
		}
		currentRequestMethod := "grpc://" + s.currentServiceName + "/" + strings.ReplaceAll(strings.Trim(method, "/"), "/", ".")
		// 目的地当前只支持grpc
		if requestChains, ok := md[HeaderRequestChain]; ok {
			for _, requestMethod := range requestChains {
				if requestMethod == currentRequestMethod {
					requestChains = append(requestChains, currentRequestMethod)
					return status.Errorf(codes.Aborted, "cycle call, chains: %s", strings.Join(requestChains, " -> "))
				}
			}
			md[HeaderRequestChain] = append(md[HeaderRequestChain], currentRequestMethod)
		}
		return invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc)
	}
}
