> 主要是用来在框架初始化后加入一些逻辑来引导业务系统的启动和在系统停止后做一些清理逻辑,
> 比如框架中的gorm数据库的连接就是通过bootstrap引导连接并断开的

详细示例参考[这里](https://github.com/asjard/asjard/blob/develop/_examples/svc-example/services/user.go)

## 如何实现

实现如下方法

```go
// Initiator 初始化需要实现的方法
type Initiator interface {
	// 启动
	Start() error
	// 停止
	Stop()
}

```

### Bootstrap

> 框架初始化完成后服务启动之前执行

添加到引导队列中

```go
import "github.com/asjard/asjard/core/bootstrap"

type CustomeBootstrap struct{}

// 系统初始化后会执行如下方法
func(CustomeBootstrap) Start() error {return nil}
// 系统停止后会执行如下方法
func(CustomeBootstrap) Stop() {}

func init() {
	// 添加到启动引导队列
	bootstrap.AddBootstrap(&CustomeBootstrap{})
}
```

示例

您可以参考[https://github.com/asjard/asjard/blob/main/pkg/stores/xgorm/xgorm.go](https://github.com/asjard/asjard/blob/main/pkg/stores/xgorm/xgorm.go)

```go
// DBManager 数据连接维护
type DBManager struct {
	dbs sync.Map
}

// Bootstrap 连接到数据库
func (m *DBManager) Bootstrap() error {
	logger.Debug("store gorm start")
	conf, err := m.loadAndWatchConfig()
	if err != nil {
		return err
	}
	return m.connDBs(conf)
}

// Shutdown 和数据库断开连接
func (m *DBManager) Shutdown() {
	m.dbs.Range(func(key, value any) bool {
		conn, ok := value.(*DBConn)
		if ok {
			sqlDB, err := conn.db.DB()
			if err == nil {
				sqlDB.Close()
			}
			m.dbs.Delete(key)
		}
		return true
	})
}
```

### Initiator

> 在本地配置文件加载完成后执行

添加到初始化队列

```go

import "github.com/asjard/asjard/core/bootstrap"

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
	bootstrap.AddInitiator(clientManager)
}
```

您可以参考[etcd连接](https://github.com/asjard/asjard/blob/develop/pkg/stores/xetcd/etcd.go)实现
