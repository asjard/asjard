> Bootstrap is a feature that execute at after framework inited and before server start

## Add bootstrap

Implement this methods at below

```go
// BootstrapHandler need implement function
type BootstrapHandler interface {
	// execute at bootstrap
	Bootstrap() error
	// execute at shutdown
	Shutdown()
}
```

Add custome bootstrap

```go
import "github.com/asjard/asjard/core/bootstrap"

// custome bootstrap implement
type CustomeBootstrap struct{}
func(CustomeBootstrap) Bootstrap() error {return nil}
func(CustomeBootstrap) Shutdown() {}

func init() {
	// add custome bootstrap
	bootstrap.AddBootstrap(&CustomeBootstrap{})
}
```

Example

reference [https://github.com/asjard/asjard/blob/main/pkg/stores/xgorm/xgorm.go](https://github.com/asjard/asjard/blob/main/pkg/stores/xgorm/xgorm.go)

```go
// DBManager database manager
type DBManager struct {
	dbs sync.Map
}

// Bootstrap connect to all databases
func (m *DBManager) Bootstrap() error {
	logger.Debug("store gorm start")
	conf, err := m.loadAndWatchConfig()
	if err != nil {
		return err
	}
	return m.connDBs(conf)
}

// Shutdown disconnect from database
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
