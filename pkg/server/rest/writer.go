package rest

import (
	"net/http"
	"reflect"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/utils"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
		statusCode = response.Status
	}
	if err := writeJSON(c, int(statusCode), response); err != nil {
		logger.Error("write json fail", "err", err)
	}
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

func writeJSON(c *Context, statusCode int, body proto.Message) error {
	c.Response.Header.Set(fasthttp.HeaderContentType, MIME_JSON)
	c.Response.SetStatusCode(statusCode)
	b, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}.Marshal(body)
	if err != nil {
		return err
	}
	if _, err := c.RequestCtx.Write(b); err != nil {
		return err
	}
	return nil
}
