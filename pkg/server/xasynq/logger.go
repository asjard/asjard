package xasynq

import (
	"fmt"

	"github.com/asjard/asjard/core/logger"
)

// asynqLogger is an adapter that satisfies the asynq.Logger interface.
// It maps Asynq's logging calls to Asjard's core logging library.
type asynqLogger struct{}

// Debug logs high-volume, granular information useful for troubleshooting.
func (l *asynqLogger) Debug(args ...any) {
	logger.Debug(fmt.Sprint(args...))
}

// Info logs general operational messages.
// Note: These are mapped to Debug in this implementation to keep the main logs
// focused on business events while still allowing task details to be seen in debug mode.
func (l *asynqLogger) Info(args ...any) {
	logger.Debug(fmt.Sprint(args...))
}

// Warn logs non-critical issues that might require attention but don't stop the worker.
func (l *asynqLogger) Warn(args ...any) {
	logger.Warn(fmt.Sprint(args...))
}

// Error logs serious issues encountered during task processing or server operation.
func (l *asynqLogger) Error(args ...any) {
	logger.Error(fmt.Sprint(args...))
}

// Fatal logs critical failures. Since Asjard prefers managed lifecycles,
// this is mapped to Error to allow the framework to attempt recovery or graceful exit.
func (l *asynqLogger) Fatal(args ...any) {
	logger.Error(fmt.Sprint(args...))
}
