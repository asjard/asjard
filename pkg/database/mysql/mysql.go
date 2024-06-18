package mysql

import (
	"fmt"
	"sync"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
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
	// ConfigKey 配置key
	ConfigKey = "database.mysql"
	// ConfigOptionsKey .
	ConfigOptionsKey            = ConfigKey + ".options"
	DbConfigKey                 = ConfigKey + ".dbs"
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
	db *gorm.DB
	// 是否可以连接
	ok bool
	// 无法连接错误原因
	err error
}

// DBConf 数据库配置
type DBConf struct {
	// 数据库连接配置
	Dsn string `json:"dsn"`
	// 驱动名称
	Driver string `json:"driver"`
	// 驱动自定义配置
	Options map[string]string `json:"options"`
}

// Options .
type Options struct {
	connName string
}

// Option .
type Option func(*Options)

var dbManager *DBManager

func init() {
	dbManager = &DBManager{}
	bootstrap.AddBootstrap(dbManager)
}

// WithConnName .
func WithConnName(connName string) func(*Options) {
	return func(opt *Options) {
		opt.connName = connName
	}
}

// 连接到数据库
func (m *DBManager) Bootstrap() error {
	logger.Debug("database mysql bootstrap")
	dbConfs := make(map[string]*DBConf)
	if err := config.GetWithUnmarshal(DbConfigKey, &dbConfs, config.WithMatchWatch(ConfigKey+".*", m.watch)); err != nil {
		return err
	}
	return m.conn(dbConfs)
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
func DB(opts ...Option) (*gorm.DB, error) {
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
	if !db.ok && db.err != nil {
		return nil, status.Error(codes.Internal, db.err.Error())
	}
	return db.db, nil
}

func (m *DBManager) conn(dbConfs map[string]*DBConf) error {
	for dbName, cfg := range dbConfs {
		logger.Debug("connect to database", "database", dbName, "config", cfg)
		db, err := gorm.Open(m.dialector(dbName, cfg), &gorm.Config{
			Logger: &mysqlLogger{},
		})
		if err != nil {
			return fmt.Errorf("connect to %s fail[%s]", dbName, err.Error())
		}
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxIdleConns(config.GetInt(ConfigOptionsKey+".maxIdleConns", 10))
		sqlDB.SetMaxOpenConns(config.GetInt(ConfigOptionsKey+".maxOpenConns", 100))
		sqlDB.SetConnMaxIdleTime(config.GetDuration(ConfigOptionsKey+".connMaxIdleTime", 10*time.Second))
		sqlDB.SetConnMaxLifetime(config.GetDuration(ConfigOptionsKey+".connMaxLifeTime", 1*time.Hour))
		conn := &DBConn{
			db: db,
			ok: true,
		}
		go conn.ping()
		m.dbs.Store(dbName, conn)
	}
	return nil
}

func (m *DBManager) dialector(name string, cfg *DBConf) gorm.Dialector {
	switch cfg.Driver {
	case postgresDefaultDriverName:
		return postgres.New(postgres.Config{
			DriverName: config.GetString(fmt.Sprintf("%s.%s.options.driverName", ConfigKey, name), ""),
			DSN:        cfg.Dsn,
		})
	case sqliteDefaultDriverName:
		return sqlite.New(sqlite.Config{
			DriverName: config.GetString(fmt.Sprintf("%s.%s.options.driverName", ConfigKey, name), ""),
			DSN:        cfg.Dsn,
		})
	case sqlserverDefaultDrierName:
		return sqlserver.New(sqlserver.Config{
			DriverName: config.GetString(fmt.Sprintf("%s.%s.options.driverName", ConfigKey, name), ""),
			DSN:        cfg.Dsn,
		})
	case clickhouseDefaultDriverName:
		return clickhouse.New(clickhouse.Config{
			DriverName: config.GetString(fmt.Sprintf("%s.%s.options.driverName", ConfigKey, name), ""),
			DSN:        cfg.Dsn,
		})
	default:
		return mysql.New(mysql.Config{
			DriverName: config.GetString(fmt.Sprintf("%s.%s.options.driverName", ConfigKey, name), ""),
			DSN:        cfg.Dsn,
		})
	}
}

func (m *DBManager) watch(event *config.Event) {
}

func (m *DBConn) ping() {
	for {
		sqlDB, err := m.db.DB()
		if err != nil {
			m.err = err
			m.ok = false
		}
		if err := sqlDB.Ping(); err != nil {
			m.err = err
			m.ok = false
		}
		time.Sleep(10 * time.Second)
	}
}

func defaultOptions() *Options {
	return &Options{
		connName: defaultConnectName,
	}
}
