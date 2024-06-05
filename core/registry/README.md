> 服务注册和发现相关功能, 服务发现和注册拆分成了两个功能，可单独实现，也可合并实现

## 服务注册

- 将本服务注册到不同的服务注册中心
- 服务注册将发生在服务启动之后

### 自定义服务注册

需实现如下所有方法

```go
// Register 服务注册相关功能
type Register interface {
	// 将服务注册到不同的配置中心
	// 如果开启心跳，则每隔一个心跳间隔注册一次
	Registe(instance *server.Instance) error
	// 从配置中心移除服务实例
	Remove(instance *server.Instance)
	// 注册中心名称
	Name() string
}
```

然后在init方法中调用`registry.AddRegister方法`, 例如

```go
func init() {
  registry.AddRegister(NewCustomeRegister)
}

// 自定义服务发现
type CustomeRegister struct{}

func NewCustomeRegister() (registry.Register, error) {
  return &CustomeRegister{}, nil
}

// 将服务注册到服务中心
func(c CustomeRegister) Registe(instance *server.Instance) error {
  // TODO 注册逻辑
  return nil
}

// 从注册中心删除服务
func (c CustomeRegister) Remove(instance *server.Instance) {
  // TODO 从注册中心删除该实例
}

func (c CustomeRegister) Name() string{
  return "custome_register"
}
```

## 服务发现

- 从不同的服务发现中心发现服务并维护在本地
- 提供相关的接口从本地查询相关服务列表


### 自定义服务发现

需实现如下所有接口

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

然后在init方法中调用`registry.AddDiscover`, 例如

```go
func init() {
	// 添加本地服务发现
	AddDiscover(NewLocalDiscover)
}
// LocalRegistry 本地服务发现
type LocalRegistry struct {
	cb func(event *Event)
	// key: serviceName
	instances               []*Instance
	localDiscoverConfPrefix string
	dm                      sync.RWMutex
}

// NewLocalDiscover .
func NewLocalDiscover() (Discovery, error) {
	localDiscover := &LocalRegistry{
		localDiscoverConfPrefix: "registry.localDiscover",
	}
	localDiscover.getAndWatch()
	return localDiscover, nil
}

// GetAll 获取所有服务列表
func (l *LocalRegistry) GetAll() ([]*Instance, error) {
	return l.instances, nil
}

// Watch 监听配置变化
func (l *LocalRegistry) Watch(callback func(event *Event)) {
	l.cb = callback
}

// Name 返回本地注册中心名称
func (l *LocalRegistry) Name() string {
	return LocalRegistryName
}

func (l *LocalRegistry) getAndWatch() {
	services := make(map[string][]string)
	if err := config.GetWithUnmarshal(l.localDiscoverConfPrefix,
		&services,
		config.WithMatchWatch(l.localDiscoverConfPrefix+".*", l.watch)); err != nil {
		logger.Errorf("get registry.localDiscover fail[%s]", err.Error())
	} else {
		l.instances = l.getInstances(services)
	}

}
func (l *LocalRegistry) watch(event *config.Event) {
	services := make(map[string][]string)
	config.GetWithUnmarshal(l.localDiscoverConfPrefix, &services)
	instances := l.getInstances(services)
	for _, instance := range l.instances {
		l.cb(&Event{
			Type:     EventTypeDelete,
			Instance: instance,
		})
	}

	for _, instance := range instances {
		l.cb(&Event{
			Type:     EventTypeUpdate,
			Instance: instance,
		})

	}
	l.instances = instances
}

func (l *LocalRegistry) getInstances(services map[string][]string) []*Instance {
	var instances []*Instance
	for name, addresses := range services {
		instance := server.NewInstance()
		instance.Name = name
		endpoints := make(map[string][]string)
		for index := range addresses {
			u, err := url.Parse(addresses[index])
			if err == nil {
				endpoints[u.Scheme] = append(endpoints[u.Scheme], u.Host)
			}
		}
		for protocol, addresses := range endpoints {
			if err := instance.AddEndpoints(protocol, map[string][]string{
				constant.ServerListenAddressName: addresses,
			}); err != nil {
				logger.Errorf(err.Error())
			}
		}
		instances = append(instances, &Instance{
			DiscoverName: l.Name(),
			Instance:     instance,
		})
	}
	return instances
}
```
