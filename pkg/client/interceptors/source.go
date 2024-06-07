package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/runtime"
	"google.golang.org/grpc"
)

const (
	// HeaderSourceServiceName 源服务名称
	HeaderSourceServiceName = "x-request-source"
	// HeaderSourceMethod 源服务方法
	HeaderSourceMethod = "x-request-method"
)

// SourceInterceptor 来源拦截器
type SourceInterceptor struct {
	currentServiceName string
}

func init() {
	client.AddInterceptor(NewSourceInterceptor)
}

// NewSourceInterceptor 初始化来源拦截器
func NewSourceInterceptor() client.ClientInterceptor {
	return &SourceInterceptor{
		currentServiceName: runtime.APP + "/" + runtime.Name,
	}
}

// Name 拦截器名称
func (SourceInterceptor) Name() string {
	return "sourceInterceptor"
}

// Interceptor 拦截器
// 上下文中添加当前服务
// 如果出现循环服务则拦截
func (s SourceInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc grpc.ClientConnInterface, invoker client.UnaryInvoker) error {
		// methodName := strings.ReplaceAll(strings.Trim(method, "/"), "/", ".")
		// logger.Debug("add source header", "method", methodName)
		// // 上游服务请求头
		// upstreamHeader := make(map[string][]string)
		// if restCtx, ok := ctx.(*rest.Context); ok {
		// 	// 上游是rest服务
		// 	upstreamHeader = restCtx.ReadHeaderParams()
		// } else {
		// 	logger.Debug("---------upstream is grpc---------")
		// 	upstreamHeader, _ = metadata.FromIncomingContext(ctx)
		// }

		// // 判断是否循环
		// serviceNameExist := false
		// if values, ok := upstreamHeader[HeaderSourceServiceName]; ok {
		// 	for _, name := range values {
		// 		if name == s.currentServiceName {
		// 			serviceNameExist = true
		// 			break
		// 		}
		// 	}
		// }
		// methodExist := false
		// if values, ok := upstreamHeader[HeaderSourceMethod]; ok && serviceNameExist {
		// 	for _, name := range values {
		// 		if name == method {
		// 			methodExist = true
		// 			break
		// 		}
		// 	}
		// }

		// if serviceNameExist && methodExist {
		// 	return status.Errorf(http.StatusLoopDetected, "loop call")
		// }

		// var nextCtx context.Context
		// // 写入请求头
		// if _, ok := cc.(*grpc.ClientConn); ok {
		// 	// 客户端是grpc
		// 	nextCtx = metadata.AppendToOutgoingContext(ctx, HeaderSourceServiceName, s.currentServiceName)
		// 	nextCtx = metadata.AppendToOutgoingContext(nextCtx, HeaderSourceMethod, methodName)
		// } else {
		// 	nextCtx = ctx
		// }
		return invoker(ctx, method, req, reply, cc)
	}
}
