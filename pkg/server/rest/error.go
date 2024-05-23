package rest

import (
	"net/http"
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
		c.writeJSON(http.StatusOK, status)
		return
	}
	// 非预定义的错误
	c.writeJSON(http.StatusOK, ErrInterServerError)
}
