package rest

import (
	"net/http"
	"reflect"
	"sync"

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
	// HeaderResponseRequestMethod 请求方法返回头
	HeaderResponseRequestMethod = "x-request-method"
	// HeaderResponseRequestID 请求ID返回头
	HeaderResponseRequestID = "x-request-id"
	// 默认writer名称
	DefaultWriterName = "default"
)

// Writer 结果输出
type Writer func(ctx *Context, data any, err error)

var (
	writers = map[string]Writer{
		DefaultWriterName: DefaultWriter,
	}
	wm sync.RWMutex
)

// AddWriter 添加writer
func AddWriter(name string, writer Writer) {
	wm.Lock()
	writers[name] = writer
	wm.Unlock()
}

// GetWriter 获取writer
func GetWriter(name string) Writer {
	wm.RLock()
	defer wm.RUnlock()
	if name == "" {
		return writers[DefaultWriterName]
	}
	w, ok := writers[name]
	if ok {
		return w
	}
	return writers[DefaultWriterName]
}

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
	st.RequestId = string(c.Response.Header.Peek(HeaderResponseRequestID))
	st.RequestMethod = string(c.Response.Header.Peek(HeaderResponseRequestMethod))
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
	if _, err := c.Write(b); err != nil {
		return err
	}
	return nil
}
