package xgorm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	ajutils "github.com/asjard/asjard/utils"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// xgormLogger implements gormLogger.Interface to intercept GORM's internal events
// and route them through the Asjard structured logging system.
type xgormLogger struct {
	logLevel gormLogger.LogLevel
	name     string // Database connection name

	ignoreRecordNotFoundError bool          // Whether to suppress errors for 'Record Not Found'
	slowThreshold             time.Duration // Execution time above which a query is logged as WARN
	slogger                   *logger.Logger
	m                         sync.RWMutex // Protects config changes during hot-reloads
}

// loggerConfig defines the schema for database-specific logging settings in the config files.
type loggerConfig struct {
	logger.Config                                  // Embedded base logging config (Level, Format, etc.)
	IgnoreRecordNotFoundError bool                 `json:"ignoreRecordNotFoundError"`
	SlowThreshold             ajutils.JSONDuration `json:"slowThreshold"`
}

var (
	// Ensure xgormLogger matches the expected GORM interface.
	_ gormLogger.Interface = &xgormLogger{}

	defaultConfig = loggerConfig{
		Config:                    logger.DefaultConfig,
		IgnoreRecordNotFoundError: true, // Default to true to reduce noise in common logic
	}
	// Cache loggers by database name to prevent redundant initializations.
	dbLoggers sync.Map
)

// NewLogger initializes a new GORM logger or returns an existing one.
// It sets up the configuration watcher for real-time updates.
func NewLogger(name string) (gormLogger.Interface, error) {
	value, ok := dbLoggers.Load(name)
	if ok {
		return value.(*xgormLogger), nil
	}
	nlogger := &xgormLogger{
		name: name,
	}
	// Initial load and setup of the config center listener.
	if err := nlogger.loadAndWatch(); err != nil {
		return nil, err
	}
	dbLoggers.Store(name, nlogger)
	return nlogger, nil
}

// LogMode clones the logger with a specific level (used by GORM for session-based overrides).
func (l *xgormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	l.m.RLock()
	defer l.m.RUnlock()
	return &xgormLogger{
		logLevel:                  level,
		ignoreRecordNotFoundError: l.ignoreRecordNotFoundError,
		slowThreshold:             l.slowThreshold,
		name:                      l.name,
		slogger:                   l.slogger,
	}
}

// Info logs standard database informational messages.
func (l *xgormLogger) Info(ctx context.Context, format string, v ...any) {
	l.m.RLock()
	defer l.m.RUnlock()
	l.slogger.L(ctx).Info(fmt.Sprintf(format, v...), "db", l.name)
}

// Warn logs database warning messages.
func (l *xgormLogger) Warn(ctx context.Context, format string, v ...any) {
	l.m.RLock()
	defer l.m.RUnlock()
	l.slogger.L(ctx).Warn(fmt.Sprintf(format, v...), "db", l.name)
}

// Error logs database error messages.
func (l *xgormLogger) Error(ctx context.Context, format string, v ...any) {
	l.m.RLock()
	defer l.m.RUnlock()
	l.slogger.L(ctx).Error(fmt.Sprintf(format, v...), "db", l.name)
}

// Trace is the core method of GORM logging. It captures SQL execution details,
// timing, and error state for every query performed by the ORM.
func (l *xgormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	l.m.RLock()
	defer l.m.RUnlock()

	elapsed := time.Since(begin)
	sql, rows := fc() // Execute the closure to get the SQL string and row count

	switch {
	// 1. Log Errors: if an error occurred and it's not a suppressed 'Not Found' error.
	case err != nil && l.logLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		l.slogger.L(ctx).Error(err.Error(),
			"sql", sql,
			"row", rows,
			"line", utils.FileWithLineNum(), // Path to the Go code that triggered this SQL
			"cost", elapsed.String(),
			"db", l.name)

	// 2. Log Slow Queries: if the duration exceeds the configured threshold.
	case elapsed > l.slowThreshold && l.slowThreshold != 0:
		l.slogger.L(ctx).Warn(fmt.Sprintf("SLOW SQL >= %s", l.slowThreshold.String()),
			"sql", sql,
			"row", rows,
			"line", utils.FileWithLineNum(),
			"cost", elapsed.String(),
			"db", l.name)

	// 3. Log Debug/Info: prints all SQL statements if the log level is set high enough.
	case l.logLevel == gormLogger.Info:
		l.slogger.L(ctx).Debug(sql,
			"row", rows,
			"line", utils.FileWithLineNum(),
			"cost", elapsed.String(),
			"db", l.name)
	}
}

// loadAndWatch attaches a listener to the config center for "asjard.logger" changes.
func (l *xgormLogger) loadAndWatch() error {
	if err := l.load(); err != nil {
		return err
	}
	config.AddPrefixListener("asjard.logger", l.watch)
	return nil
}

// load parses configuration settings into the logger struct and resets the underlying slogger.
func (l *xgormLogger) load() error {
	conf := defaultConfig
	// Load base logger settings
	if err := config.GetWithUnmarshal("asjard.logger", &conf.Config); err != nil {
		return err
	}
	// Load specific GORM overrides
	if err := config.GetWithUnmarshal("asjard.logger.gorm", &conf); err != nil {
		return err
	}

	l.m.Lock()
	defer l.m.Unlock()

	// Initialize the structured logger. WithCallerSkip(5) is used to ensure the log
	// shows the business logic file/line, not the logger's internal wrapper.
	l.slogger = logger.DefaultLogger(slog.New(logger.NewSlogHandler(&conf.Config))).WithCallerSkip(5)
	l.ignoreRecordNotFoundError = conf.IgnoreRecordNotFoundError
	l.slowThreshold = conf.SlowThreshold.Duration
	return nil
}

// watch triggers whenever the configuration center detects a change in "asjard.logger".
func (l *xgormLogger) watch(event *config.Event) {
	if err := l.load(); err != nil {
		logger.Error("gorm watch config fail", "err", err)
	}
}
