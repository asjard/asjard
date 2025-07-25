package xgorm

import (
	"context"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	config.Set("asjard.stores.gorm.dbs.default.dsn", "test_default.db")
	config.Set("asjard.stores.gorm.dbs.default.driver", "sqlite")

	config.Set("asjard.stores.gorm.dbs.another.dsn", "test_another.db")
	config.Set("asjard.stores.gorm.dbs.another.driver", "sqlite")

	config.Set("asjard.stores.gorm.dbs.ciphered.dsn", "dGVzdF9jaXBoZXJlZC5kYg==")
	config.Set("asjard.stores.gorm.dbs.ciphered.driver", "sqlite")
	config.Set("asjard.stores.gorm.dbs.ciphered.cipherName", "base64")

	config.Set("asjard.stores.gorm.dbs.auto_decrypt.dsn", "encrypted_base64:dGVzdF9jaXBoZXJlZC5kYg==")
	config.Set("asjard.stores.gorm.dbs.auto_decrypt.driver", "sqlite")

	config.Set("asjard.stores.gorm.dbs.lock.dsn", "root:my-secret-pw@tcp(127.0.0.1:3306)/example-database?charset=utf8&parseTime=True&loc=Local&timeout=5s&readTimeout=5s")
	config.Set("asjard.stores.gorm.dbs.lock.driver", "mysql")
	time.Sleep(50 * time.Millisecond)
	if err := dbManager.Start(); err != nil {
		panic(err)
	}
	m.Run()
	dbManager.Stop()

}

func TestLoadAndWatchConfig(t *testing.T) {
	conf, err := dbManager.loadAndWatchConfig()
	assert.Nil(t, err)
	assert.Equal(t, 5, len(conf))
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
	t.Run("cipher", func(t *testing.T) {
		db, err := DB(context.Background(), WithConnName("ciphered"))
		assert.Nil(t, err)
		assert.NotNil(t, db)
		err = db.AutoMigrate(&testTable{})
		assert.Nil(t, err)
		err = db.Create(&testTable{DBName: "cipher"}).Error
		assert.Nil(t, err)
		var result testTable
		err = db.Where("db_name=?", "cipher").First(&result).Error
		assert.Nil(t, err)
		assert.NotEmpty(t, result.DBName)
	})
	t.Run("auto_decrpt", func(t *testing.T) {
		db, err := DB(context.Background(), WithConnName("auto_decrypt"))
		assert.Nil(t, err)
		assert.NotNil(t, db)
		err = db.AutoMigrate(&testTable{})
		assert.Nil(t, err)
		err = db.Create(&testTable{DBName: "auto_decrypt"}).Error
		assert.Nil(t, err)
		var result testTable
		err = db.Where("db_name=?", "auto_decrypt").First(&result).Error
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
		time.Sleep(100 * time.Millisecond)
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
