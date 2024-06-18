package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/asjard/asjard/core/logger"
	gormLogger "gorm.io/gorm/logger"
)

type mysqlLogger struct{}

var _ gormLogger.Interface = &mysqlLogger{}

func (mysqlLogger) LogMode(gormLogger.LogLevel) gormLogger.Interface {
	return &mysqlLogger{}
}

func (mysqlLogger) Info(ctx context.Context, format string, v ...any) {
	logger.Info(fmt.Sprintf(format, v...))
}

func (mysqlLogger) Warn(ctx context.Context, format string, v ...any) {
	logger.Warn(fmt.Sprintf(format, v...))
}
func (mysqlLogger) Error(ctx context.Context, format string, v ...any) {
	logger.Error(fmt.Sprintf(format, v...))
}
func (mysqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {

}
