package status

import (
	"fmt"
	"net/http"
)

// Status .
type Status struct {
	// 错误码, 因为会包含HTTP状态码
	// 非http.StatusText范围内的错误码为自定义错误码
	// 自定义错误码格式为: XXXY
	// 其中:
	// 	XXX: 表示HTTP状态码，固定三位数字
	//  Y: 表示自定义错误码，位数不限制
	Code uint32 `json:"code"`
	// 错误信息
	Message string `json:"message"`
	// 错误处理文档
	Doc string `json:"doc"`
}

// StatusOption .
type StatusOption func(*Status)

var (
	// StatusInternalServerError 系统内部错误状态
	StatusInternalServerError = &Status{
		Code:    http.StatusInternalServerError,
		Message: "Internal Server Error",
	}
	// ErrNotFound .
	ErrNotFound = Errorf(http.StatusNotFound, "Page Not Found")
	// ErrMethodNotAllowed .
	ErrMethodNotAllowed = Errorf(http.StatusMethodNotAllowed, "Method Not Allowed")
	// ErrInterServerError .
	ErrInterServerError = Errorf(http.StatusInternalServerError, "Internal Server Error")
)

// Error .
func (s *Status) Error() string {
	if s.Code == 0 {
		return ""
	}
	return fmt.Sprintf("(%d)%s,doc:%s", s.Code, s.Message, s.Doc)
}

// Errorf returns an error representing code, msg and doc.  If code is 0, returns nil.
// 非http.StatusText范围内的错误码为自定义错误码
// 自定义错误码格式为: XXXY
// 其中:
//
//	XXX: 表示HTTP状态码，固定三位数字
//	Y: 表示自定义错误码，位数不限制
func Errorf(code uint32, msg string, options ...StatusOption) error {
	if code == 0 {
		return nil
	}
	status := &Status{Code: code, Message: msg}
	for _, opt := range options {
		opt(status)
	}
	return status
}
