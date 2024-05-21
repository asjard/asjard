package logger

import (
	"github.com/asjard/asjard/core/logger"
	"go.uber.org/zap"
)

// ZapLogger zap日志
type ZapLogger struct {
	*zap.SugaredLogger
}

var _ logger.Logger = &ZapLogger{}

func init() {
	logger.AddLogger(New)
}

// New 初始化日志框架
func New() (logger.Logger, error) {
	lg, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{SugaredLogger: lg.Sugar()}, nil
}

// Info .
func (l *ZapLogger) Info(v ...any) {
	l.SugaredLogger.Info(v...)
}

// Infof .
func (l *ZapLogger) Infof(format string, v ...any) {
	l.SugaredLogger.Infof(format, v...)
}

// Debug .
func (l *ZapLogger) Debug(v ...any) {
	l.SugaredLogger.Debug(v...)
}

// Debugf .
func (l *ZapLogger) Debugf(format string, v ...any) {
	l.SugaredLogger.Debugf(format, v...)
}

// Warn .
func (l *ZapLogger) Warn(v ...any) {
	l.SugaredLogger.Warn(v...)
}

// Warnf .
func (l *ZapLogger) Warnf(format string, v ...any) {
	l.SugaredLogger.Warnf(format, v...)
}

// Error .
func (l *ZapLogger) Error(v ...any) {
	l.SugaredLogger.Error(v...)
}

// Errorf .
func (l *ZapLogger) Errorf(format string, v ...any) {
	l.SugaredLogger.Errorf(format, v...)
}

// SetLevel 设置日志级别
func (l *ZapLogger) SetLevel(level logger.Level) {

}
