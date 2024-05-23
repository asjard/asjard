package rest

import "github.com/asjard/asjard/core/logger"

// Logger .
type Logger struct{}

// Printf .
func (Logger) Printf(format string, args ...any) {
	logger.Infof(format, args...)
}
