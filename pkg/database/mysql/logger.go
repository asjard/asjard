package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/asjard/asjard/core/logger"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type mysqlLogger struct {
	logLevel                  gormLogger.LogLevel
	ignoreRecordNotFoundError bool
	slowThreshold             time.Duration
	name                      string
}

var _ gormLogger.Interface = &mysqlLogger{}

func (l *mysqlLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return &mysqlLogger{
		logLevel: level,
	}
}

func (l mysqlLogger) Info(ctx context.Context, format string, v ...any) {
	logger.Info(fmt.Sprintf(format, v...), "db", l.name)
}

func (l mysqlLogger) Warn(ctx context.Context, format string, v ...any) {
	logger.Warn(fmt.Sprintf(format, v...), "db", l.name)
}
func (l mysqlLogger) Error(ctx context.Context, format string, v ...any) {
	logger.Error(fmt.Sprintf(format, v...), "db", l.name)
}
func (l mysqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
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
