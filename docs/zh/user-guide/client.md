## 全局配置

> 其他自定义客户端如果没有配置如下字段，则继承全局配置

```yaml
## 客户端相关配置
asjard:
  clients:
    ## 客户端负载均衡, 不存在或者为空，则默认为localityRoundRobin
    # loadbalance: "localityRoundRobin"
    ## 自定义客户端拦截器配置
    ## 同servers.interceptors配置
    # interceptors: ""
    ## 框架内建客户端拦截器
    # builtInInterceptors: errLog,slowLog,rest2RpcContext,cycleChainInterceptor,circuitBreaker
    ## 或者可以按照yaml列表配置
    # builtInInterceptors:
    #   - rest2RpcContext
    #   - cycleChainInterceptor
    #   - circuitBreaker
    ## 同servers.certFile配置
    # certFile: ""
    ## grpc客户端相关配置
```

## 自定义客户端

实现如下方法

```go
// ClientConnInterface 客户端需要实现的接口
// 对grpc.ClientConnInterface扩展
type ClientConnInterface interface {
	grpc.ClientConnInterface
	// 客户端连接的服务名称
	ServiceName() string
	// 客户端连接的协议
	Protocol() string
	Conn() grpc.ClientConnInterface
}
```
