package xhttp

import (
	"bytes"

	"github.com/asjard/asjard/core/logger"
)

// logWriter adapts net/http.Server.ErrorLog to the Asjard logger.
type logWriter struct{}

// Write implements io.Writer for *log.Logger.
func (logWriter) Write(p []byte) (int, error) {
	logger.Error(string(bytes.TrimSpace(p)))
	return len(p), nil
}
