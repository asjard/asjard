package xgorm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
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

var _ gormLogger.Interface = &xgormLogger{}

func NewLogger(name string, ignoreRecordNotFoundError bool, slowThreshold time.Duration) (gormLogger.Interface, error) {
	lg := &xgormLogger{
		ignoreRecordNotFoundError: ignoreRecordNotFoundError,
		slowThreshold:             slowThreshold,
		name:                      name,
	}
	if err := lg.loadAndWatch(); err != nil {
		return nil, err
	}
	return lg, nil
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
	conf := logger.DefaultConfig
	if err := config.GetWithUnmarshal("asjard.logger.gorm", &conf); err != nil {
		return err
	}
	l.slogger = slog.New(logger.NewSlogHandler(conf))
	return nil
}

func (l *xgormLogger) watch(event *config.Event) {
	if err := l.load(); err != nil {
		logger.Error("gorm watch config fail", "err", err)
	}
}
