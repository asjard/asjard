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

type xgormLogger struct {
	logLevel gormLogger.LogLevel
	name     string

	ignoreRecordNotFoundError bool
	slowThreshold             time.Duration
	// slogger                   *slog.Logger
	slogger *logger.Logger
	m       sync.RWMutex
}

type loggerConfig struct {
	logger.Config
	IgnoreRecordNotFoundError bool                 `json:"ignoreRecordNotFoundError"`
	SlowThreshold             ajutils.JSONDuration `json:"slowThreshold"`
}

var (
	_ gormLogger.Interface = &xgormLogger{}

	defaultConfig = loggerConfig{
		Config:                    logger.DefaultConfig,
		IgnoreRecordNotFoundError: true,
	}
	dbLoggers sync.Map
)

// NewLogger 日志初始化
func NewLogger(name string) (gormLogger.Interface, error) {
	value, ok := dbLoggers.Load(name)
	if ok {
		return value.(*xgormLogger), nil
	}
	nlogger := &xgormLogger{
		name: name,
	}
	if err := nlogger.loadAndWatch(); err != nil {
		return nil, err
	}
	dbLoggers.Store(name, nlogger)
	return nlogger, nil
}

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

func (l *xgormLogger) Info(ctx context.Context, format string, v ...any) {
	l.m.RLock()
	defer l.m.RUnlock()
	l.slogger.L(ctx).Info(fmt.Sprintf(format, v...), "db", l.name)
}

func (l *xgormLogger) Warn(ctx context.Context, format string, v ...any) {
	l.m.RLock()
	defer l.m.RUnlock()
	l.slogger.L(ctx).Warn(fmt.Sprintf(format, v...), "db", l.name)
}
func (l *xgormLogger) Error(ctx context.Context, format string, v ...any) {
	l.m.RLock()
	defer l.m.RUnlock()
	l.slogger.L(ctx).Error(fmt.Sprintf(format, v...), "db", l.name)
}
func (l *xgormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	l.m.RLock()
	defer l.m.RUnlock()
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && l.logLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		l.slogger.L(ctx).Error(err.Error(), "sql", sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	case elapsed > l.slowThreshold && l.slowThreshold != 0:
		l.slogger.L(ctx).Warn(fmt.Sprintf("SLOW SQL >= %s", l.slowThreshold.String()), "sql", sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	case l.logLevel == gormLogger.Info:
		l.slogger.L(ctx).Debug(sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	}
}

func (l *xgormLogger) loadAndWatch() error {
	if err := l.load(); err != nil {
		return err
	}
	config.AddPrefixListener("asjard.logger", l.watch)
	return nil
}

func (l *xgormLogger) load() error {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.logger", &conf.Config); err != nil {
		return err
	}
	if err := config.GetWithUnmarshal("asjard.logger.gorm", &conf); err != nil {
		return err
	}
	l.m.Lock()
	defer l.m.Unlock()
	// l.slogger = slog.New(logger.NewSlogHandler(&conf.Config))
	l.slogger = logger.DefaultLogger(slog.New(logger.NewSlogHandler(&conf.Config))).WithCallerSkip(5)
	l.ignoreRecordNotFoundError = conf.IgnoreRecordNotFoundError
	l.slowThreshold = conf.SlowThreshold.Duration
	return nil
}

func (l *xgormLogger) watch(event *config.Event) {
	if err := l.load(); err != nil {
		logger.Error("gorm watch config fail", "err", err)
	}
}
