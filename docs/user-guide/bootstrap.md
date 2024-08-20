> 主要是用来在框架初始化后加入一些逻辑来引导业务系统的启动和在系统停止后做一些清理逻辑,
> 比如框架中的gorm数据库的连接就是通过bootstrap引导连接并断开的

## 如何实现

实现如下方法

```go
// BootstrapHandler 启动引导需实现的方法
type BootstrapHandler interface {
	// 启动时执行
	Bootstrap() error
	// 停止时执行
	Shutdown()
}
```

并添加到引导队列中

```go
import "github.com/asjard/asjard/core/bootstrap"

type CustomeBootstrap struct{}

// 系统初始化后会执行如下方法
func(CustomeBootstrap) Bootstrap() error {return nil}
// 系统停止后会执行如下方法
func(CustomeBootstrap) Shutdown() {}

func init() {
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
