package xgorm

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const (
	testSourceName     = "testSource"
	testSourcePriority = 0
)

type testSource struct {
	cb      func(*config.Event)
	configs map[string]any
	cm      sync.RWMutex
}

func newTestSource() (config.Sourcer, error) {
	return &testSource{
		configs: map[string]any{
			"asjard.stores.gorm.dbs.default.dsn":    "test_default.db",
			"asjard.stores.gorm.dbs.default.driver": "sqlite",

			"asjard.stores.gorm.dbs.another.dsn":    "test_another.db",
			"asjard.stores.gorm.dbs.another.driver": "sqlite",

			"asjard.config.setDefaultSource": testSourceName,
		},
	}, nil
}

func (s *testSource) GetAll() map[string]*config.Value {
	configs := make(map[string]*config.Value)
	for key, value := range s.configs {
		configs[key] = &config.Value{
			Sourcer: s,
			Value:   value,
		}
	}
	return configs
}

// 添加配置到配置源中
func (s *testSource) Set(key string, value any) error {
	s.cm.Lock()
	s.configs[key] = value
	s.cm.Unlock()
	s.cb(&config.Event{
		Type: config.EventTypeCreate,
		Key:  key,
		Value: &config.Value{
			Sourcer: s,
			Value:   value,
		},
	})
	return nil
}

// 监听配置变化,当配置源中的配置发生变化时,
// 通过此回调方法通知config_manager进行配置变更
func (s *testSource) Watch(callback func(event *config.Event)) error {
	s.cb = callback
	return nil
}

// 和配置中心断开连接
func (s *testSource) Disconnect() {}

// 配置中心的优先级
func (s *testSource) Priority() int { return testSourcePriority }

// 配置源名称
func (s *testSource) Name() string { return testSourceName }

func initTestConfig() {
	if err := config.AddSource(testSourceName, testSourcePriority, newTestSource); err != nil {
		panic(err)
	}
	if err := config.Load(-1); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	initTestConfig()
	if err := dbManager.Start(); err != nil {
		panic(err)
	}
	m.Run()
	dbManager.Stop()

}

func TestLoadAndWatchConfig(t *testing.T) {
	conf, err := dbManager.loadAndWatchConfig()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(conf))
}

type testTable struct {
	gorm.Model
	DBName string `gorm:"column:db_name"`
}

func TestConnDBs(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		db, err := DB(context.Background())
		assert.Nil(t, err)
		assert.NotNil(t, db)
		err = db.AutoMigrate(&testTable{})
		assert.Nil(t, err)
		err = db.Create(&testTable{DBName: "default"}).Error
		assert.Nil(t, err)
		var result testTable
		err = db.Where("db_name=?", "default").First(&result).Error
		assert.Nil(t, err)
		assert.NotEmpty(t, result.DBName)
	})
	t.Run("another", func(t *testing.T) {
		db, err := DB(context.Background(), WithConnName("another"))
		assert.Nil(t, err)
		assert.NotNil(t, db)
		err = db.AutoMigrate(&testTable{})
		assert.Nil(t, err)
		err = db.Create(&testTable{DBName: "another"}).Error
		assert.Nil(t, err)
		var result testTable
		err = db.Where("db_name=?", "another").First(&result).Error
		assert.Nil(t, err)
		assert.NotEmpty(t, result.DBName)
		var result1 testTable
		err = db.Where("db_name=?", "default").First(&result1).Error
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Empty(t, result1.DBName)
	})
	t.Run("newdb", func(t *testing.T) {
		config.Set("asjard.stores.gorm.dbs.newdb.dsn", "test_new.db")
		config.Set("asjard.stores.gorm.dbs.newdb.driver", "sqlite")
		// 设置配置是异步过程，等待数据库连接刷新
		time.Sleep(200 * time.Millisecond)
		db, err := DB(context.Background(), WithConnName("newdb"))
		assert.Nil(t, err)
		assert.NotNil(t, db)
	})
	t.Run("shutdown", func(t *testing.T) {
		dbManager.Stop()
		_, err := DB(context.TODO())
		assert.NotNil(t, err)
	})
}
