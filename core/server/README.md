> 服务实现，例如grpc，http


## 服务实现

需实现如下方法:

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
	ListenAddresses() map[string]string
	// 是否已启用
	Enabled() bool
}
```

然后调用`AddServer`方法

可参考`pkg/server/rest`


## 拦截器实现

需实现如下方法:

```go
// ServerInterceptor 服务拦截器需要实现的方法
type ServerInterceptor interface {
	// 拦截器名称
	Name() string
	// 拦截器
	Interceptor() UnaryServerInterceptor
}
```

然后调用`AddInterceptor`添加拦截器

可参考`pkg/server/rest/interceptor.go`
