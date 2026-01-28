> 将本服务注册到注册中心，或者从服务发现中心发现服务

## 已实现服务发现

- [x] [local](registry-local.md)
- [x] [etcd](registry-etcd.md)
- [x] [consul](registry-consul.md)

## 已实现服务注册

- [x] [etcd](registry-etcd.md)
- [x] [consul](registry-consul.md)

## 配置

```yaml
asjard:
  ## 服务发现，注册相关配置
  registry:
    ## 是否自动服务注册
    # autoRegiste: true
    ## 延迟注册, 服务启动后等待延迟时间后注册服务到注册中心
    # delayRegiste: 0s
    ## 注册心跳, 开启后每隔一个心跳时间注册服务到服务注册中心
    # heartbeat: false
    ## 心跳频率
    # heartbeatInterval: 5s

    ## 自动服务发现, 自动从配置中心发现服务
    # autoDiscove: true
    ## 服务健康检查，检查服务是否正常，如果不正常则从本地缓存中删除该服务
    # healthCheck: false
    ## 健康检查间隔时间
    # healthCheckInterval: 10s
    ## 认定检查失败的检查阈值(连续失败次数)
    # failureThreshold: 1
```

## 使用

```go
package main

import _ "github.com/asjard/asjard/pkg/registry/etcd"
```

## 自定义服务注册

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

## 自定义服务发现

> 实现如下方法

```go
// Discovery 服务发现相关功能
type Discovery interface {
	// 获取所有服务实例
	GetAll() ([]*Instance, error)
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

## 服务发现

```go
import "github.com/asjard/asjard/core/registry"

func (api *GwAPI) GetServiceInstances(ctx context.Context, in *pb.ServiceInstancesReq) (*pb.ServiceInstancesResp, error) {
	var instances []*pb.ServiceInstancesResp_Instance
	for _, item := range registry.PickServices(registry.WithServiceName(in.ServiceName)) {
		instances = append(instances, &pb.ServiceInstancesResp_Instance{
			ServiceName: item.Service.Instance.Name,
			InstanceId:  item.Service.Instance.ID,
		})
	}

	return &pb.ServiceInstancesResp{
		Instances: instances,
	}, nil
}
```

你也可以通过registry库中的`WithXXX`方法选择指定服务
