package rest

import "sync"

// Response 请求返回
type Response struct {
	*Status
	// 请求数据
	Data any `json:"data"`
}

var responsePool = sync.Pool{
	New: func() any {
		return &Response{
			Status: &Status{},
			Data:   nil,
		}
	},
}

func newResponse(c *Context, data any, err error) *Response {
	response := responsePool.Get().(*Response)
	if err == nil {
		response.Status = &Status{}
	} else {
		if status, ok := err.(*Status); ok {
			response.Status = status
		} else {
			response.Status = StatusInternalServerError
		}
	}
	if response.Status.Code != 0 && response.Status.Doc == "" {
		response.Status.Doc = c.errPage
	}
	response.Data = data
	c.response = response
	return response
}
