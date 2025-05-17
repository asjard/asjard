package xasynq

import (
	"fmt"

	"github.com/asjard/asjard/core/logger"
)

type asynqLogger struct{}

func (l *asynqLogger) Debug(args ...any) {
	logger.Debug(fmt.Sprint(args...))
}
func (l *asynqLogger) Info(args ...any) {
	logger.Debug(fmt.Sprint(args...))
}
func (l *asynqLogger) Warn(args ...any) {
	logger.Warn(fmt.Sprint(args...))
}
func (l *asynqLogger) Error(args ...any) {
	logger.Error(fmt.Sprint(args...))
}
func (l *asynqLogger) Fatal(args ...any) {
	logger.Error(fmt.Sprint(args...))
}
