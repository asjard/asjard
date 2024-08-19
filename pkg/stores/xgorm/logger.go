package xgorm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	ajutils "github.com/asjard/asjard/utils"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type xgormLogger struct {
	logLevel gormLogger.LogLevel

	ignoreRecordNotFoundError bool
	slowThreshold             time.Duration
	name                      string

	slogger *slog.Logger
}

type loggerConfig struct {
	*logger.Config
	IgnoreRecordNotFoundError bool                 `json:"ignoreRecordNotFoundError"`
	SlowThreshold             ajutils.JSONDuration `json:"slowThreshold"`
}

var (
	_ gormLogger.Interface = &xgormLogger{}

	defaultConfig = loggerConfig{
		Config:                    logger.DefaultConfig,
		IgnoreRecordNotFoundError: true,
	}

	glogger *xgormLogger
)

// InitLogger 日志初始化
func InitLogger() error {
	lg := &xgormLogger{}
	if err := lg.loadAndWatch(); err != nil {
		return err
	}
	glogger = lg
	return nil
}

func NewLogger(name string) gormLogger.Interface {
	return &xgormLogger{
		ignoreRecordNotFoundError: glogger.ignoreRecordNotFoundError,
		slowThreshold:             glogger.slowThreshold,
		name:                      name,
		slogger:                   glogger.slogger,
	}
}

func (l *xgormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return &xgormLogger{
		logLevel:                  level,
		ignoreRecordNotFoundError: l.ignoreRecordNotFoundError,
		slowThreshold:             l.slowThreshold,
		name:                      l.name,
		slogger:                   l.slogger,
	}
}

func (l *xgormLogger) Info(ctx context.Context, format string, v ...any) {
	l.slogger.Info(fmt.Sprintf(format, v...), "db", l.name)
}

func (l *xgormLogger) Warn(ctx context.Context, format string, v ...any) {
	l.slogger.Warn(fmt.Sprintf(format, v...), "db", l.name)
}
func (l *xgormLogger) Error(ctx context.Context, format string, v ...any) {
	l.slogger.Error(fmt.Sprintf(format, v...), "db", l.name)
}
func (l *xgormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && l.logLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		l.slogger.Error(err.Error(), "sql", sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	case elapsed > l.slowThreshold && l.slowThreshold != 0:
		l.slogger.Warn(fmt.Sprintf("SLOW SQL >= %s", l.slowThreshold.String()), "sql", sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	case l.logLevel == gormLogger.Info:
		l.slogger.Debug(sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	}
}

func (l *xgormLogger) loadAndWatch() error {
	if err := l.load(); err != nil {
		return err
	}
	config.AddPrefixListener("asjard.logger.gorm", l.watch)
	return nil
}

func (l *xgormLogger) load() error {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.logger.gorm", &conf); err != nil {
		return err
	}
	l.slogger = slog.New(logger.NewSlogHandler(conf.Config))
	l.ignoreRecordNotFoundError = conf.IgnoreRecordNotFoundError
	l.slowThreshold = conf.SlowThreshold.Duration
	return nil
}

func (l *xgormLogger) watch(event *config.Event) {
	if err := l.load(); err != nil {
		logger.Error("gorm watch config fail", "err", err)
	}
}
