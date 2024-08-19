> 将本服务注册到注册中心，或者从服务发现中心发现服务

### 已实现服务发现

- [x] local
- [x] etcd
- [x] consul

### 已实现服务注册

- [x] etcd
- [x] consul

### 本地服务发现

#### 配置

```yaml
asjard:
  registry:
    localDiscover:
      ## 服务名称
      helloGrpc:
        ## 服务列表, 格式: {protocol}://{ip}:{port}
        - grpc://127.0.0.1:7010
```

### etcd服务注册发现

#### 配置

```yaml
asjard:
  registry:
    etcd:
      ## 客户端名称, asjard.database.etcd.clients.{这里的名称}
      client: default
```

#### 使用

```go
package main

import _ "github.com/asjard/asjard/pkg/registry/etcd"
```

### 自定义服务注册

> 实现如下方法

```go
// Register 服务注册相关功能
type Register interface {
	// 将服务注册到不同的配置中心
	// 如果开启心跳，则每隔一个心跳间隔注册一次
	Registe(instance *server.Service) error
	// 从配置中心移除服务实例
	Remove(instance *server.Service)
	// 注册中心名称
	Name() string
}
```

注册

```go
// 注册
func init() {
	registry.AddRegister("custome_register_name", NewRegisterFunc)
}
```

使用

```go
package main

import _ "you_custome_register_dir"

```

### 自定义服务发现

> 实现如下方法

```go
// Discovery 服务发现相关功能
type Discovery interface {
	// 获取所有服务实例
	GetAll() ([]*Instance, error)
	// 监听服务变化
	Watch(callbak func(event *Event))
	// 服务发现中心名称
	Name() string
}
```

注册

```go
func init() {
	registry.AddDiscover("your_custome_discover_name", NewDiscoverFunc)
}
```

使用

```go
package main

import _ "your_custome_discover_dir"
```
