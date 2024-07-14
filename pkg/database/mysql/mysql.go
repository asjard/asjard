package mysql

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/metrics"
	"github.com/asjard/asjard/utils"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

const (
	postgresDefaultDriverName   = "postgres"
	mysqlDefaultDriverName      = "mysql"
	sqliteDefaultDriverName     = "sqlite"
	sqlserverDefaultDrierName   = "sqlserver"
	clickhouseDefaultDriverName = "clickhouse"
	defaultConnectName          = "default"
)

// DBManager 数据连接维护
type DBManager struct {
	dbs sync.Map
}

// DBConn 数据库连接
type DBConn struct {
	name  string
	db    *gorm.DB
	debug bool
}

// Config 数据库配置
type Config struct {
	Dbs     map[string]DBConnConfig `json:"dbs"`
	Options Options                 `json:"options"`
}

type Options struct {
	MaxIdleConns              int                `json:"maxIdleConns"`
	MaxOpenConns              int                `json:"maxOpenConns"`
	ConnMaxIdleTime           utils.JSONDuration `json:"connMaxIdleTime"`
	ConnMaxLifeTime           utils.JSONDuration `json:"connMaxLifeTime"`
	Debug                     bool               `json:"debug"`
	IgnoreRecordNotFoundError bool               `json:"ignoreRecordNotFoundError"`
	SlowThreshold             utils.JSONDuration `json:"slowThreshold"`
}

// DBConnConfig 数据库连接配置
type DBConnConfig struct {
	// 数据库连接配置
	Dsn string `json:"dsn"`
	// 驱动名称
	Driver string `json:"driver"`
	// 驱动自定义配置
	Options DBConnOptions `json:"options"`
}

// DBConnOptions 数据库连接自定义配置
type DBConnOptions struct {
	CustomeDriverName string `json:"driverName"`
}

// DBOptions .
type DBOptions struct {
	connName string
}

// Option .
type Option func(*DBOptions)

var dbManager *DBManager

func init() {
	dbManager = &DBManager{}
	bootstrap.AddBootstrap(dbManager)
}

// WithConnName .
func WithConnName(connName string) func(*DBOptions) {
	return func(opt *DBOptions) {
		opt.connName = connName
	}
}

// 连接到数据库
func (m *DBManager) Bootstrap() error {
	logger.Debug("database mysql bootstrap")
	cfg := Config{
		Dbs: make(map[string]DBConnConfig),
		Options: Options{
			MaxIdleConns:              10,
			MaxOpenConns:              100,
			ConnMaxIdleTime:           utils.JSONDuration{Duration: 10 * time.Second},
			ConnMaxLifeTime:           utils.JSONDuration{Duration: time.Hour},
			Debug:                     false,
			IgnoreRecordNotFoundError: true,
			SlowThreshold:             utils.JSONDuration{Duration: 200 * time.Millisecond},
		},
	}
	if err := config.GetWithUnmarshal(constant.ConfigDatabaseMysqlPrefix,
		&cfg,
		config.WithMatchWatch(constant.ConfigDatabaseMysqlPrefix+".*", m.watch)); err != nil {
		return err
	}
	return m.conn(cfg)
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
		}
		return true
	})
}

// DB 数据库连接地址
func DB(ctx context.Context, opts ...Option) (*gorm.DB, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	conn, ok := dbManager.dbs.Load(options.connName)
	if !ok {
		return nil, status.Error(codes.Internal, "database not found")
	}
	db, ok := conn.(*DBConn)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid db")
	}
	if db.debug {
		return db.db.Debug().WithContext(ctx), nil
	}
	return db.db.WithContext(ctx), nil
}

func (m *DBManager) conn(dbCfg Config) error {
	for dbName, cfg := range dbCfg.Dbs {
		logger.Debug("connect to database", "database", dbName, "config", cfg)
		db, err := gorm.Open(m.dialector(cfg), &gorm.Config{
			Logger: &mysqlLogger{
				ignoreRecordNotFoundError: dbCfg.Options.IgnoreRecordNotFoundError,
				slowThreshold:             dbCfg.Options.SlowThreshold.Duration,
				name:                      dbName,
			},
		})
		if err != nil {
			return fmt.Errorf("connect to %s fail[%s]", dbName, err.Error())
		}
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxIdleConns(dbCfg.Options.MaxIdleConns)
		sqlDB.SetMaxOpenConns(dbCfg.Options.MaxOpenConns)
		sqlDB.SetConnMaxIdleTime(dbCfg.Options.ConnMaxIdleTime.Duration)
		sqlDB.SetConnMaxLifetime(dbCfg.Options.ConnMaxLifeTime.Duration)
		conn := &DBConn{
			name:  dbName,
			db:    db,
			debug: dbCfg.Options.Debug,
		}
		m.dbs.Store(dbName, conn)
		metrics.RegisterCollector("db_"+dbName, collectors.NewDBStatsCollector(sqlDB, dbName))
	}
	return nil
}

func (m *DBManager) dialector(cfg DBConnConfig) gorm.Dialector {
	switch cfg.Driver {
	case postgresDefaultDriverName:
		return postgres.New(postgres.Config{
			DriverName: cfg.Options.CustomeDriverName,
			DSN:        cfg.Dsn,
		})
	case sqliteDefaultDriverName:
		return sqlite.New(sqlite.Config{
			DriverName: cfg.Options.CustomeDriverName,
			DSN:        cfg.Dsn,
		})
	case sqlserverDefaultDrierName:
		return sqlserver.New(sqlserver.Config{
			DriverName: cfg.Options.CustomeDriverName,
			DSN:        cfg.Dsn,
		})
	case clickhouseDefaultDriverName:
		return clickhouse.New(clickhouse.Config{
			DriverName: cfg.Options.CustomeDriverName,
			DSN:        cfg.Dsn,
		})
	default:
		return mysql.New(mysql.Config{
			DriverName: cfg.Options.CustomeDriverName,
			DSN:        cfg.Dsn,
		})
	}
}

func (m *DBManager) watch(event *config.Event) {
}

func defaultOptions() *DBOptions {
	return &DBOptions{
		connName: defaultConnectName,
	}
}
