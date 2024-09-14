package xasynq

import (
	"github.com/asjard/asjard/core/logger"
)

type asynqLogger struct{}

func (l *asynqLogger) Debug(args ...any) {
	logger.Debug("asynq", args...)
}
func (l *asynqLogger) Info(args ...any) {
	logger.Info("asynq", args...)
}
func (l *asynqLogger) Warn(args ...any) {
	logger.Warn("asynq", args...)
}
func (l *asynqLogger) Error(args ...any) {
	logger.Error("asynq", args...)
}
func (l *asynqLogger) Fatal(args ...any) {
	logger.Error("asynq", args...)
}
