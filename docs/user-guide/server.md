> 可以通过暴露已有的服务，或者自行实现相应的协议队外暴露服务

### 已实现服务

- [x] [rest](./server-rest.md)
- [x] [grpc](./server-grpc.md)
- [x] [pprof](./server-pprof.md)

### 如何实现自己的服务

#### 配置约定

- 配置应都放在`asjard.servers.{自定义服务名称}`该命名空间下

#### 自定义服务

> 具体可参考`core/server/server.go`中的代码

- 需实现如下两个方法

```go

// Server 每个协议需要实现的内容
type Server interface {
	// 注册
	AddHandler(handler any) error
	// 服务启动
	Start(startErr chan error) error
	// 服务停止
	Stop()
	// 服务提供的协议
	Protocol() string
	// 服务监听地址列表
	// key为监听地址名称, listen,advertise为保留关键词，会在客户端负载均衡场景中用到
	// value为监听地址
	ListenAddresses() map[string]string
	// 是否已启用
	Enabled() bool
}

// NewServerFunc 服务初始化方法
type NewServerFunc func(options *ServerOptions) (Server, error)
```

- 然后通过`AddServer`添加服务

```go
func init() {
	server.AddServer(Protocol, New)
}
```

- 具体实现可参考`pkg/server/rest`,`pkg/server/grpc`,`pkg/server/pprof`
