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
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/glebarez/sqlite"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"

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
	// DefaultConnectName is the default key used to store the primary database connection.
	DefaultConnectName = "default"
)

// DBManager maintains a thread-safe registry of active GORM database connections.
type DBManager struct {
	// dbs is a map of connection name (string) to *DBConn.
	dbs sync.Map

	// cm protects access to the raw configs map used for monitoring changes.
	cm      sync.RWMutex
	configs map[string]*DBConnConfig
}

// DBConn wraps the GORM DB instance with metadata like its name and debug status.
type DBConn struct {
	name  string
	db    *gorm.DB
	debug bool
}

// Config represents the top-level configuration structure for database stores.
type Config struct {
	DBs     map[string]DBConnConfig `json:"dbs"`
	Options Options                 `json:"options"`
}

// Options defines global behavioral and performance settings for all database connections.
type Options struct {
	MaxIdleConns              int                `json:"maxIdleConns"`
	MaxOpenConns              int                `json:"maxOpenConns"`
	ConnMaxIdleTime           utils.JSONDuration `json:"connMaxIdleTime"`
	ConnMaxLifeTime           utils.JSONDuration `json:"connMaxLifeTime"`
	Debug                     bool               `json:"debug"`
	SkipInitializeWithVersion bool               `json:"skipInitializeWithVersion"`
	// Traceable: When true, enables OpenTelemetry tracing for SQL operations.
	Traceable bool `json:"traceable"`
	// Metricsable: When true, exports database connection pool statistics to Prometheus.
	Metricsable bool `json:"metricsable"`

	// Standard GORM configuration flags
	SkipDefaultTransaction                   bool `json:"skipDefaultTransaction"`
	FullSaveAssociations                     bool `json:"fullSaveAssociations"`
	DryRun                                   bool `json:"dryRun"`
	DisableAutomaticPing                     bool `json:"disableAutomaticPing"`
	PrepareStmt                              bool `json:"prepareStmt"`
	DisableForeignKeyConstraintWhenMigrating bool `json:"disableForeignKeyConstraintWhenMigrating"`
	IgnoreRelationshipsWhenMigrating         bool `json:"ignoreRelationshipsWhenMigrating"`
	DisableNestedTransaction                 bool `json:"disableNestedTransaction"`
	AllowGlobalUpdate                        bool `json:"allowGlobalUpdate"`
	QueryFields                              bool `json:"queryFields"`
	CreateBatchSize                          int  `json:"createBatchSize"`
	TranslateError                           bool `json:"translateError"`
	PropagateUnscoped                        bool `json:"propagateUnscoped"`
}

// DBConnConfig holds the specific connection details for a single database cluster.
type DBConnConfig struct {
	// Dsn is the connection string (Data Source Name).
	Dsn string `json:"dsn"`
	// CipherName allows the DSN to be stored as an encrypted string for security.
	CipherName   string         `json:"cipherName"`
	CipherParams map[string]any `json:"cipherParams"`
	// Driver defines which SQL driver to use (e.g., mysql, postgres).
	Driver string `json:"driver"`
	// Options contains per-connection overrides for global settings.
	Options DBConnOptions `json:"options"`
}

// DBConnOptions combines generic options with driver-specific naming overrides.
type DBConnOptions struct {
	Options
	CustomeDriverName string `json:"driverName"`
}

// DBOptions used for functional options pattern when fetching a client.
type DBOptions struct {
	connName string
}

type Option func(*DBOptions)

// ctxDBKey is an unexported type for context keys to avoid collisions with
// other packages. This ensures that only this package can access the DB
// instance stored in the context.
type ctxDBKey struct{}

// WithConnName allows the caller to request a specific named database connection.
func WithConnName(connName string) func(*DBOptions) {
	return func(opts *DBOptions) {
		if connName != "" {
			opts.connName = connName
		}
	}
}

// WithDB wraps the parent context and injects a *gorm.DB instance using a private context key.
// This allows downstream functions to retrieve the DB connection (or an active transaction)
// from the context, ensuring consistency across different layers of the application.
// Example:
//
//	db, err := xgorm.DB(ctx)
//	if err != nil {
//		return err
//	}
//	return db.Transaction(func(tx *gorm.DB) error {
//		// inject tx into ctx
//		anotherFn(xgorm.WithDB(ctx, tx))
//	})
func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxDBKey{}, db)
}

var (
	dbManager *DBManager
	// Sensible framework defaults for database pooling.
	defaultConnOptions = Options{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxIdleTime: utils.JSONDuration{Duration: 10 * time.Second},
		ConnMaxLifeTime: utils.JSONDuration{Duration: time.Hour},
		QueryFields:     true,
		PrepareStmt:     true,
	}
)

func init() {
	dbManager = &DBManager{configs: make(map[string]*DBConnConfig)}
	// Registers as a bootstrap component to initialize DBs during startup.
	bootstrap.AddBootstrap(dbManager)
}

// DB retrieves an established GORM database connection from the manager.
// It automatically injects the context and configures debug mode if required.
// If no database is found in the context, it returns the global manager's instance
func DB(ctx context.Context, opts ...Option) (*gorm.DB, error) {
	if db, ok := ctx.Value(ctxDBKey{}).(*gorm.DB); ok {
		return db, nil
	}

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
	// Apply debug mode and context to the GORM session.
	if db.debug {
		return db.db.Debug().WithContext(ctx), nil
	}
	return db.db.WithContext(ctx), nil
}

// NewDB creates a fresh database connection bypasses the registry cache.
func NewDB(ctx context.Context, opts ...Option) (*gorm.DB, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	dbManager.cm.RLock()
	connConf, ok := dbManager.configs[options.connName]
	dbManager.cm.RUnlock()
	if !ok {
		logger.Error("db not found", "db", options.connName)
		return nil, status.DatabaseNotFoundError()
	}
	db, err := dbManager.connDB(options.connName, connConf)
	if err != nil {
		return nil, err
	}
	if connConf.Options.Debug {
		return db.Debug().WithContext(ctx), nil
	}
	return db.WithContext(ctx), nil
}

// Start initiates the database manager by loading configurations and establishing initial connections.
func (m *DBManager) Start() error {
	logger.Debug("gorm start")
	conf, err := m.loadAndWatchConfig()
	if err != nil {
		return err
	}
	return m.connDBs(conf)
}

// Stop closes all active database connections gracefully.
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

// connDBs establishes physical connections for all provided database configurations.
func (m *DBManager) connDBs(dbsConf map[string]*DBConnConfig) error {
	for dbName, cfg := range dbsConf {
		db, err := m.connDB(dbName, cfg)
		if err != nil {
			logger.Error("connect to database fail", "database", dbName, "config", cfg, "err", err)
			return err
		}
		m.dbs.Store(dbName, &DBConn{
			name:  dbName,
			db:    db,
			debug: cfg.Options.Debug,
		})
		logger.Debug("connect to database success", "database", dbName, "config", cfg)
	}
	return nil
}

// connDB handles individual GORM initialization, including logging, pooling, tracing, and metrics.
func (m *DBManager) connDB(dbName string, cfg *DBConnConfig) (*gorm.DB, error) {
	// Initialize the custom structured logger for this DB instance.
	dbLogger, err := NewLogger(dbName)
	if err != nil {
		return nil, err
	}
	// Select the driver dialector based on the driver name.
	dial, err := m.dialector(cfg)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dial, &gorm.Config{
		SkipDefaultTransaction:                   cfg.Options.SkipDefaultTransaction,
		FullSaveAssociations:                     cfg.Options.FullSaveAssociations,
		DryRun:                                   cfg.Options.DryRun,
		DisableAutomaticPing:                     cfg.Options.DisableAutomaticPing,
		PrepareStmt:                              cfg.Options.PrepareStmt,
		DisableForeignKeyConstraintWhenMigrating: cfg.Options.DisableForeignKeyConstraintWhenMigrating,
		IgnoreRelationshipsWhenMigrating:         cfg.Options.IgnoreRelationshipsWhenMigrating,
		DisableNestedTransaction:                 cfg.Options.DisableNestedTransaction,
		AllowGlobalUpdate:                        cfg.Options.AllowGlobalUpdate,
		QueryFields:                              cfg.Options.QueryFields,
		CreateBatchSize:                          cfg.Options.CreateBatchSize,
		TranslateError:                           cfg.Options.TranslateError,
		PropagateUnscoped:                        cfg.Options.PropagateUnscoped,
		Logger:                                   dbLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to %s fail[%s]", dbName, err.Error())
	}
	// Inject OpenTelemetry tracing middleware if enabled.
	if cfg.Options.Traceable {
		db.Use(tracing.NewPlugin(tracing.WithDBName(dbName), tracing.WithoutMetrics()))
	}

	// Obtain the standard library sql.DB to configure the connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.Options.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Options.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(cfg.Options.ConnMaxIdleTime.Duration)
	sqlDB.SetConnMaxLifetime(cfg.Options.ConnMaxLifeTime.Duration)

	// Register Prometheus collectors for monitoring database stats.
	if cfg.Options.Metricsable {
		metrics.RegisterCollector("db_"+dbName+"_collector", collectors.NewDBStatsCollector(sqlDB, dbName))
	}
	return db, nil
}

// dialector creates a GORM dialector for the requested driver, handling potential DSN decryption.
func (m *DBManager) dialector(cfg *DBConnConfig) (gorm.Dialector, error) {
	dsn := cfg.Dsn
	// If a cipher is specified, decrypt the connection string before passing it to the driver.
	if cfg.CipherName != "" {
		plainDsn, err := security.Decrypt(dsn, security.WithCipherName(cfg.CipherName), security.WithParams(cfg.CipherParams))
		if err != nil {
			return nil, err
		}
		dsn = plainDsn
	}
	// Map driver strings to their respective GORM driver initializers.
	switch cfg.Driver {
	case postgresDefaultDriverName:
		return postgres.New(postgres.Config{DriverName: cfg.Options.CustomeDriverName, DSN: dsn}), nil
	case sqliteDefaultDriverName:
		return sqlite.Open(dsn), nil
	case sqlserverDefaultDrierName:
		return sqlserver.New(sqlserver.Config{DriverName: cfg.Options.CustomeDriverName, DSN: dsn}), nil
	default:
		return mysql.New(mysql.Config{
			DriverName:                cfg.Options.CustomeDriverName,
			DSN:                       dsn,
			SkipInitializeWithVersion: cfg.Options.SkipInitializeWithVersion,
		}), nil
	}
}

// loadAndWatchConfig loads initial DB settings and subscribes to remote configuration changes.
func (m *DBManager) loadAndWatchConfig() (map[string]*DBConnConfig, error) {
	conf, err := m.loadConfig()
	if err != nil {
		return conf, err
	}
	config.AddPatternListener("asjard.stores.gorm.*", m.watch)
	return conf, nil
}

// loadConfig unmarshals database and global options from the configuration center.
func (m *DBManager) loadConfig() (map[string]*DBConnConfig, error) {
	dbs := make(map[string]*DBConnConfig)
	options := defaultConnOptions
	if err := config.GetWithUnmarshal("asjard.stores.gorm.options", &options); err != nil {
		return dbs, err
	}
	if err := config.GetWithUnmarshal("asjard.stores.gorm.dbs", &dbs); err != nil {
		return dbs, err
	}
	// Merge global options with individual DB overrides.
	for dbName, dbConfig := range dbs {
		dbConfig.Options.Options = options
		if err := config.GetWithUnmarshal(fmt.Sprintf("asjard.stores.gorm.dbs.%s.options", dbName),
			&dbConfig.Options.Options); err != nil {
			logger.Error("load gorm db options fail", "database", dbName, "err", err)
		}
	}
	m.cm.Lock()
	m.configs = dbs
	m.cm.Unlock()
	return dbs, nil
}

// watch handles real-time configuration updates by reconnecting updated DBs and removing deleted ones.
func (m *DBManager) watch(event *config.Event) {
	conf, err := m.loadConfig()
	if err != nil {
		logger.Error("load gorm config fail", "err", err)
		return
	}
	// Re-connect or update existing connections.
	if err := m.connDBs(conf); err != nil {
		logger.Error("connect db fail", "err", err)
		return
	}
	// Clean up connections that were removed from the updated configuration.
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
