package rest

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/asjard/asjard/utils"
	"github.com/valyala/fasthttp"
)

const (
	// QueryParamNeedStatusCode 需要http状态码query请求参数
	QueryParamNeedStatusCode = "nsc"
)

// DefaultWriter 默认输出
// 当data和err都为nil约定为已自行write
func DefaultWriter(c *Context, data any, err error) {
	if err == nil && (data == nil || reflect.ValueOf(data).IsNil()) {
		return
	}
	response := newResponse(c, data, err)
	var statusCode uint32 = http.StatusOK
	if c.URI().QueryArgs().Has(QueryParamNeedStatusCode) {
		statusCode = getStatusCode(response.Code)
	}
	writeJSON(c, int(statusCode), response)
	response.Status = nil
	response.Data = nil
	responsePool.Put(response)
}

func getStatusCode(code uint32) uint32 {
	if code != 0 {
		if http.StatusText(int(code)) != "" {
			return code
		}
		if code < 1000 {
			return http.StatusInternalServerError
		}
		return code / utils.Uint32Len(code)
	}
	return http.StatusOK
}

func writeJSON(c *Context, statusCode int, body any) error {
	if body == nil {
		c.Response.SetStatusCode(statusCode)
		return nil
	}
	c.Response.Header.Set(fasthttp.HeaderContentType, MIME_JSON)
	c.Response.SetStatusCode(statusCode)
	return json.NewEncoder(c.Response.BodyWriter()).Encode(body)
}
