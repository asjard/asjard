package xgorm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/metrics"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

const (
	postgresDefaultDriverName   = "postgres"
	mysqlDefaultDriverName      = "mysql"
	sqliteDefaultDriverName     = "sqlite"
	sqlserverDefaultDrierName   = "sqlserver"
	clickhouseDefaultDriverName = "clickhouse"
	DefaultConnectName          = "default"
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
	DBs     map[string]DBConnConfig `json:"dbs"`
	Options Options                 `json:"options"`
}

// Options 数据库连接全局配置
type Options struct {
	MaxIdleConns              int                `json:"maxIdleConns"`
	MaxOpenConns              int                `json:"maxOpenConns"`
	ConnMaxIdleTime           utils.JSONDuration `json:"connMaxIdleTime"`
	ConnMaxLifeTime           utils.JSONDuration `json:"connMaxLifeTime"`
	Debug                     bool               `json:"debug"`
	SkipInitializeWithVersion bool               `json:"skipInitializeWithVersion"`
	SkipDefaultTransaction    bool               `json:"skipDefaultTransaction"`
	TranslateError            bool               `json:"translateError"`
	// 是否开启链路追踪
	Traceable bool `json:"traceable"`
	// 是否开启监控
	Metricsable bool `json:"metricsable"`
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
	Options
	CustomeDriverName string `json:"driverName"`
}

// DBOptions .
type DBOptions struct {
	connName string
}

// Option .
type Option func(*DBOptions)

// WithConnName .
func WithConnName(connName string) func(*DBOptions) {
	return func(opts *DBOptions) {
		if connName != "" {
			opts.connName = connName
		}
	}
}

var (
	dbManager          *DBManager
	defaultConnOptions = Options{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxIdleTime: utils.JSONDuration{Duration: 10 * time.Second},
		ConnMaxLifeTime: utils.JSONDuration{Duration: time.Hour},
		TranslateError:  true,
	}
)

func init() {
	dbManager = &DBManager{}
	bootstrap.AddBootstrap(dbManager)
}

// DB 数据库连接地址
func DB(ctx context.Context, opts ...Option) (*gorm.DB, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	conn, ok := dbManager.dbs.Load(options.connName)
	if !ok {
		logger.Error("db not found", "db", options.connName)
		return nil, status.DatabaseNotFoundError()
	}
	db, ok := conn.(*DBConn)
	if !ok {
		logger.Error("invalid db type, must be *DBConn", "current", fmt.Sprintf("%T", conn))
		return nil, status.InternalServerError()
	}
	if db.debug {
		return db.db.Debug().WithContext(ctx), nil
	}
	return db.db.WithContext(ctx), nil
}

// Start 连接到数据库
func (m *DBManager) Start() error {
	logger.Debug("store gorm start")
	conf, err := m.loadAndWatchConfig()
	if err != nil {
		return err
	}
	return m.connDBs(conf)
}

// Stop 和数据库断开连接
func (m *DBManager) Stop() {
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

func (m *DBManager) connDBs(dbsConf map[string]*DBConnConfig) error {
	for dbName, cfg := range dbsConf {
		if err := m.connDB(dbName, cfg); err != nil {
			logger.Error("connect to database fail", "database", dbName, "config", cfg, "err", err)
			return err
		}
		logger.Debug("connect to database success", "database", dbName, "config", cfg)
	}
	return nil
}

func (m *DBManager) connDB(dbName string, cfg *DBConnConfig) error {
	dbLogger, err := NewLogger(dbName)
	if err != nil {
		return err
	}
	db, err := gorm.Open(m.dialector(cfg), &gorm.Config{
		SkipDefaultTransaction: cfg.Options.SkipDefaultTransaction,
		TranslateError:         cfg.Options.TranslateError,
		Logger:                 dbLogger,
	})
	if err != nil {
		return fmt.Errorf("connect to %s fail[%s]", dbName, err.Error())
	}
	if cfg.Options.Traceable {
		db.Use(tracing.NewPlugin(tracing.WithDBName(dbName), tracing.WithoutMetrics()))
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(cfg.Options.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Options.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(cfg.Options.ConnMaxIdleTime.Duration)
	sqlDB.SetConnMaxLifetime(cfg.Options.ConnMaxLifeTime.Duration)
	conn := &DBConn{
		name:  dbName,
		db:    db,
		debug: cfg.Options.Debug,
	}
	m.dbs.Store(dbName, conn)
	if cfg.Options.Metricsable {
		metrics.RegisterCollector("db_"+dbName+"_collector", collectors.NewDBStatsCollector(sqlDB, dbName))
	}
	return nil
}

func (m *DBManager) dialector(cfg *DBConnConfig) gorm.Dialector {
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
			DriverName:                cfg.Options.CustomeDriverName,
			DSN:                       cfg.Dsn,
			SkipInitializeWithVersion: cfg.Options.SkipInitializeWithVersion,
		})
	default:
		return mysql.New(mysql.Config{
			DriverName:                cfg.Options.CustomeDriverName,
			DSN:                       cfg.Dsn,
			SkipInitializeWithVersion: cfg.Options.SkipInitializeWithVersion,
		})
	}
}

func (m *DBManager) loadAndWatchConfig() (map[string]*DBConnConfig, error) {
	conf, err := m.loadConfig()
	if err != nil {
		return conf, err
	}
	config.AddPatternListener("asjard.stores.gorm.*", m.watch)
	return conf, nil
}

func (m *DBManager) loadConfig() (map[string]*DBConnConfig, error) {
	dbs := make(map[string]*DBConnConfig)
	options := defaultConnOptions
	if err := config.GetWithUnmarshal("asjard.stores.gorm.options", &options); err != nil {
		return dbs, err
	}
	if err := config.GetWithUnmarshal("asjard.stores.gorm.dbs", &dbs); err != nil {
		return dbs, err
	}
	for dbName, dbConfig := range dbs {
		dbConfig.Options.Options = options
		if err := config.GetWithUnmarshal(fmt.Sprintf("asjard.database.gorm.dbs.%s.options", dbName),
			&dbConfig.Options.Options); err != nil {
			return dbs, err
		}
	}
	return dbs, nil
}

func (m *DBManager) watch(event *config.Event) {
	conf, err := m.loadConfig()
	if err != nil {
		logger.Error("load gorm config fail", "err", err)
		return
	}
	if err := m.connDBs(conf); err != nil {
		logger.Error("connect db fail", "err", err)
		return
	}
	// 删除被删除的数据库
	m.dbs.Range(func(key, value any) bool {
		if _, ok := conf[key.(string)]; !ok {
			logger.Debug("gorm remove db", "db", key)
			m.dbs.Delete(key)
		}
		return true
	})
}

func defaultOptions() *DBOptions {
	return &DBOptions{
		connName: DefaultConnectName,
	}
}
