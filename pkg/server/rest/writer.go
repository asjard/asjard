package rest

import (
	"net/http"
	"reflect"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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
	st := status.FromError(err)
	var statusCode uint32 = http.StatusOK
	if c.URI().QueryArgs().Has(QueryParamNeedStatusCode) {
		statusCode = st.Status
	}
	if err == nil {
		if d, err := anypb.New(data.(proto.Message)); err == nil {
			st.Data = d
		} else {
			logger.Error("can not create anypb.Any", "data", data, "err", err)
		}

	}
	if err := writeJSON(c, int(statusCode), st); err != nil {
		logger.Error("write json fail", "err", err)
	}
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
