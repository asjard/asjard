package rest

import (
	"fmt"

	"github.com/asjard/asjard/core/logger"
)

// Logger is a wrapper struct that satisfies the logger interface
// required by the underlying fasthttp server.
type Logger struct{}

// Printf implements the standard logger interface.
// It receives logs from the fasthttp server, formats them,
// and redirects them to the Asjard core logger at the Error level.
func (Logger) Printf(format string, args ...any) {
	// We use fmt.Sprintf to construct the final message from the server
	// and pass it to the centralized error logger.
	logger.Error(fmt.Sprintf(format, args...))
}
