package xgorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/asjard/asjard/core/logger"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type xgormLogger struct {
	logLevel                  gormLogger.LogLevel
	ignoreRecordNotFoundError bool
	slowThreshold             time.Duration
	name                      string
}

var _ gormLogger.Interface = &xgormLogger{}

func (l *xgormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return &xgormLogger{
		logLevel:                  level,
		ignoreRecordNotFoundError: l.ignoreRecordNotFoundError,
		slowThreshold:             l.slowThreshold,
		name:                      l.name,
	}
}

func (l xgormLogger) Info(ctx context.Context, format string, v ...any) {
	logger.Info(fmt.Sprintf(format, v...), "db", l.name)
}

func (l xgormLogger) Warn(ctx context.Context, format string, v ...any) {
	logger.Warn(fmt.Sprintf(format, v...), "db", l.name)
}
func (l xgormLogger) Error(ctx context.Context, format string, v ...any) {
	logger.Error(fmt.Sprintf(format, v...), "db", l.name)
}
func (l xgormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && l.logLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		logger.Error(err.Error(), "sql", sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	case elapsed > l.slowThreshold && l.slowThreshold != 0:
		logger.Warn(fmt.Sprintf("SLOW SQL >= %s", l.slowThreshold.String()), "sql", sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	case l.logLevel == gormLogger.Info:
		logger.Debug(sql, "row", rows, "line", utils.FileWithLineNum(), "cost", elapsed.String(), "db", l.name)
	}
}
