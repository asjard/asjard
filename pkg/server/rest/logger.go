package rest

import (
	"fmt"

	"github.com/asjard/asjard/core/logger"
)

// Logger .
type Logger struct{}

// Printf .
func (Logger) Printf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...))
}
