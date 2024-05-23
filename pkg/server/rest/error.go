package rest

import (
	"net/http"

	"github.com/asjard/asjard/core/logger"
)

var (
	// ErrNotFound .
	ErrNotFound = Error(http.StatusNotFound,
		"Page Not Found", "")
	// ErrMethodNotAllowed .
	ErrMethodNotAllowed = Error(http.StatusMethodNotAllowed, "Method Not Allowed", "")
	// ErrInterServerError .
	ErrInterServerError = Error(http.StatusInternalServerError, "Internal Server Error", "")
)

// DefaultErrorHandler 默认错误处理
func DefaultErrorHandler(c *Context, err error) {
	if status, ok := err.(*Status); ok {
		if status.Doc == "" {
			status.Doc = c.errPage
		}
		c.writeJSON(http.StatusOK, status)
		return
	}
	logger.Debugf("get an unexpect error: %s", err.Error())
	// 非预定义的错误
	c.writeJSON(http.StatusOK, ErrInterServerError)
}
