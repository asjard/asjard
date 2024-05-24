package rest

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

var (
	// StatusInternalServerError 系统内部错误状态
	StatusInternalServerError = &Status{
		Code:    http.StatusInternalServerError,
		Message: "Internal Server Error",
	}
	// ErrNotFound .
	ErrNotFound = Error(http.StatusNotFound, "Page Not Found", "")
	// ErrMethodNotAllowed .
	ErrMethodNotAllowed = Error(http.StatusMethodNotAllowed, "Method Not Allowed", "")
	// ErrInterServerError .
	ErrInterServerError = Error(http.StatusInternalServerError, "Internal Server Error", "")
)

// Error .
func (s *Status) Error() string {
	if s.Code == 0 {
		return ""
	}
	return fmt.Sprintf("(%d)%s,doc:%s", s.Code, s.Message, s.Doc)
}

// Error returns an error representing code, msg and doc.  If code is 0, returns nil.
// 非http.StatusText范围内的错误码为自定义错误码
// 自定义错误码格式为: XXXY
// 其中:
//
//	XXX: 表示HTTP状态码，固定三位数字
//	Y: 表示自定义错误码，位数不限制
func Error(code uint32, msg, doc string) error {
	if code == 0 {
		return nil
	}
	return &Status{Code: code, Message: msg, Doc: doc}
}
