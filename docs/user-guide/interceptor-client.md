## 接口实现

- 如需添加拦截器，则需实现如下功能
- 然后通过`client.AddInterceptor`注册拦截器，之后服务启动时将会通过配置调用相应的拦截器

```go

// ClientInterceptor defines the interface for a client-side interceptor.
// Implementing this allows a module to provide metadata and the actual interceptor logic.
type ClientInterceptor interface {
	// Name returns the unique identifier of the interceptor.
	Name() string
	// Interceptor returns the functional UnaryClientInterceptor.
	Interceptor() UnaryClientInterceptor
}

// UnaryClientInterceptor is a function that intercepts a unary RPC call.
// It can perform logic before and after the invoker is called, such as logging,
// tracing, or modifying request metadata.
type UnaryClientInterceptor func(ctx context.Context, method string, req, reply any, cc ClientConnInterface, invoker UnaryInvoker) error

```

## 已支持的实现

- [熔断降级](interceptor-client-circuit-breaker.md)
- [循环调用检测](interceptor-client-cycle-chain.md)
- [请求错误日志](interceptor-client-errlog.md)
- [HTTP请求头转GRPC上下文](inteceptor-client-rest2grpc.md)
- [慢日志](inteceptor-client-slowlog.md)
- [请求参数校验](inteceptor-client-validate.md)
- [panic日志](inteceptor-client-panic.md)

## 配置

```yaml
asjard:
  ## client configurations
  clients:
    ## 全局通用协议无关拦截器
    # interceptors: ""
    ## 全局内建默认拦截器
    # builtInInterceptors: panic,rest2RpcContext,validate,errLog,slowLog,cycleChainInterceptor
    ## 指定协议的配置
    grpc:
      ## grpc的拦截器
      # interceptors: ""
```

更新默认拦截器:

```yaml
asjard:
  ## client configurations
  clients:
    ## 全局通用协议无关拦截器
    # interceptors: ""
    ## 全局内建默认拦截器
    ## 局部更新语法
    ## -errLog 删除errLog拦截器
    ## +errLog:customeLog 在errLog拦截器前添加customeLog拦截器
    ## errLog+:customeLog 在errLog拦截器之后添加customeLog拦截器
    ## =errLog:customeLog 将errLog拦截器替换为customeLog拦截器
    ## customeLog 在默认拦截器尾部添加customeLog拦截器
    # builtInInterceptors: -errLog,customeLog
```

## 自定义实现

```go

package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
)

const (
	// PanicInterceptorName is the unique identifier for this interceptor.
	PanicInterceptorName = "panic"
)

func init() {
	// Register the panic recovery interceptor globally for all server protocols.
	client.AddInterceptor(PanicInterceptorName, NewPanic)
}

// Panic represents the recovery interceptor component.
type Panic struct{}

// NewPanic initializes the panic interceptor.
func NewPanic() (client.ClientInterceptor, error) {
	return &Panic{}, nil
}

// Interceptor returns the middleware function that handles panic recovery.
func (*Panic) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) (err error) {
		// Use defer to ensure the recovery logic runs even if the handler panics.
		defer func() {
			if rcv := recover(); rcv != nil {
				// 1. Collect diagnostic information about the crash.
				args := []any{
					"err", rcv, // The actual panic object/message.
					"req", req, // The request payload that triggered the panic.
					"method", method, // The endpoint being called.
					"protocol", cc.Protocol(), // gRPC or REST.
					"service", cc.ServiceName(),
					"stack", string(debug.Stack()), // The full goroutine stack trace.
				}

				// 2. Log the incident with Error level for alerting and debugging.
				logger.L(ctx).Error("request panic", args...)

				// 3. Mask the internal crash from the client by returning a
				// standardized 500 Internal Server Error.
				err = status.InternalServerError()
			}
		}()

		// Execute the next interceptor or the final business logic handler.
		return invoker(ctx, method, req, reply, cc)
	}
}

// Name returns the interceptor's unique name.
func (*Panic) Name() string {
	return PanicInterceptorName
}

```
