> 和[bootstrap](bootstrap.md)同理,不过比`bootstrap`更早执行,
> 在配置源优先级<=2执行完成后执行, 比如配置中心的连接

## 如何使用

实现如下方法

```go
// Initator 配置初始化后，其他组件初始化初始化之前需要执行的方法
type Initator interface {
	// 启动
	Start() error
	// 停止
	Stop()
}
```

并加入到初始化队列

```go

import "github.com/asjard/asjard/core/initator"

// ClientManager 客户端连接维护
type ClientManager struct {
	clients sync.Map
}

func (m *ClientManager) Start() error {
	clients, err := m.loadAndWatchConfig()
	if err != nil {
		return err
	}
	return m.newClients(clients)
}

func (m *ClientManager) Stop() {
	m.clients.Range(func(key, value any) bool {
		conn, ok := value.(*ClientConn)
		if ok {
			if err := conn.client.Close(); err != nil {
				logger.Error("close etcd client fail", "client", conn.name, "err", err)
			}
			m.clients.Delete(key)
		}
		return true
	})
}

func init() {
	clientManager = &ClientManager{}
	initator.AddInitator(clientManager)
}
```

您可以参考[etcd连接](https://github.com/asjard/asjard/blob/develop/pkg/stores/xetcd/etcd.go)实现
