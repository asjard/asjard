package rest

import (
	"fmt"
)

// Status .
type Status struct {
	// 错误码, 因为会包含HTTP状态码
	// 自定义错误码应>=600
	Code int32 `json:"code"`
	// 错误信息
	Message string `json:"message"`
	// 错误处理文档
	Doc string `json:"doc"`
}

// Error .
func (s *Status) Error() string {
	if s.Code == 0 {
		return ""
	}
	return fmt.Sprintf("(%d)%s,doc:%s", s.Code, s.Message, s.Doc)
}

// Error returns an error representing code, msg and doc.  If code is 0, returns nil.
func Error(code int32, msg, doc string) error {
	if code == 0 {
		return nil
	}
	return &Status{Code: code, Message: msg, Doc: doc}
}
