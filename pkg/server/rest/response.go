package rest

import (
	"sync"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/status"
)

// Response 请求返回
type Response struct {
	*status.Status
	// 请求数据
	Data any `json:"data"`
}

var responsePool = sync.Pool{
	New: func() any {
		return &Response{
			Status: &status.Status{},
			Data:   nil,
		}
	},
}

func newResponse(c *Context, data any, err error) *Response {
	response := responsePool.Get().(*Response)
	if err == nil {
		response.Status = &status.Status{}
	} else {
		if stts, ok := err.(*status.Status); ok {
			response.Status = stts
		} else {
			logger.Error(err.Error())
			response.Status = status.StatusInternalServerError
		}
	}
	if response.Status.Code != 0 && response.Status.Doc == "" {
		response.Status.Doc = c.errPage
	}
	response.Data = data
	c.response = response
	return response
}
