## 全局配置

> 其他自定义客户端如果没有配置如下字段，则继承全局配置

```yaml
asjard:
  ## 客户端相关配置
  ## 不实时更新
  clients:
    ## 客户端负载均衡, 不存在或者为空，则默认为roundRobin
    loadbalance: "roundRobin"
    ## 同servers.interceptors配置
    interceptors: "rest2RpcContext,cycleChainInterceptor,circuitBreaker"
    ## 同servers.certFile配置
    certFile: ""
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
