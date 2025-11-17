package rest

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	// MIME_XML .
	MIME_XML = "application/xml"
	// MIME_JSON .
	MIME_JSON = "application/json"
	// MIME_ZIP .
	MIME_ZIP = "application/zip"
	// MIME_OCTET .
	MIME_OCTET = "application/octet-stream"
)

// Context fasthttp.RequestCtx的封装
type Context struct {
	*fasthttp.RequestCtx
	errPage string
	write   Writer
}

var (
	contextPool = sync.Pool{
		New: func() any {
			return &Context{
				// errPage: config.GetString("asjard.service.website", ""),
			}
		},
	}
)

// NewContext .
func NewContext(ctx *fasthttp.RequestCtx, options ...Option) *Context {
	// c := &Context{
	// 	RequestCtx: ctx,
	// }
	c := contextPool.Get().(*Context)
	c.RequestCtx = ctx
	for _, opt := range options {
		opt(c)
	}
	return c
}

// ReadEntity 解析请求参数并序列化到entity中
// 解析顺序 query -> header -> body -> path
// 后解析的同名参数会覆盖前解析的同名参数
// post,put,patch解析body体,其余不解析
func (c *Context) ReadEntity(entity proto.Message) error {
	if entity == nil {
		return nil
	}
	return c.ReadEntityWithReaders(entity, c.DefaultEntityReaders())
}

func (c *Context) ReadEntityWithReaders(entity proto.Message, readers []*EntityReader) error {
	requestMethod := string(c.Method())
	for _, source := range readers {
		if _, ok := source.SkipMethods[requestMethod]; ok {
			continue
		}
		if err := source.Reader(entity); err != nil {
			return err
		}
	}
	return nil
}

// GetPathParam 获取路径参数
func (c *Context) GetUserParam(key string) []string {
	value := c.UserValueBytes(utils.UnsafeString2Byte(key))
	if value == nil {
		return []string{}
	}
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = utils.SafeByte2String(val)
	default:
		v = fmt.Sprintf("%v", val)
	}
	return []string{v}
}

// GetHeaderParam 获取请求头参数
func (c *Context) GetHeaderParam(key string) []string {
	v := c.Request.Header.PeekAll(key)
	s := make([]string, len(v))
	for idx, b := range v {
		s[idx] = utils.SafeByte2String(b)
	}
	return s
}

// GetQueryParam 获取query参数
func (c *Context) GetQueryParam(key string) []string {
	v := c.QueryArgs().PeekMulti(key)
	s := make([]string, len(v))
	for idx, b := range v {
		s[idx] = utils.SafeByte2String(b)
	}
	return s
}

// WriteData 请求返回
func (c *Context) WriteData(data any, err error) {
	if c.write == nil {
		DefaultWriter(c, data, err)
	} else {
		c.write(c, data, err)
	}
	c.Close()
}

// NewOutgoingContext .
func (c *Context) NewOutgoingContext() context.Context {
	return metadata.NewOutgoingContext(c, c.ReadHeaderParams())
}

// FromIncomingContext .
func (c *Context) FromIncomingContext() map[string][]string {
	return c.ReadHeaderParams()
}

// SetWriter 设置writer方法
func (c *Context) SetWriter(writer Writer) {
	c.write = writer
}

// Close .
func (c *Context) Close() {
	c.write = nil
	c.RequestCtx = nil
	contextPool.Put(c)
}

// JSONBodyParams 读取请求体
func (c *Context) JSONBodyParams() []byte {
	if bytes.Equal(c.Request.Header.ContentType(), []byte(MIME_JSON)) {
		return c.Request.Body()
	}
	return []byte{}
}

// ReadQueryParams 获取query参数
func (c *Context) ReadQueryParams() map[string][]string {
	queries := make(map[string][]string)
	c.QueryArgs().All()(func(key, value []byte) bool {
		k := utils.SafeByte2String(key)
		queries[k] = append(queries[k], utils.SafeByte2String(value))
		return true
	})
	return queries
}

// ReadHeaderParams 读取header请求参数
func (c *Context) ReadHeaderParams() map[string][]string {
	headers := make(map[string][]string)
	c.Request.Header.All()(func(key, value []byte) bool {
		k := utils.SafeByte2String(key)
		headers[k] = append(headers[k], utils.SafeByte2String(value))
		return true
	})
	return headers
}

// ReadPathParams 读取path请求参数
func (c *Context) ReadPathParams() map[string][]string {
	params := make(map[string][]string)
	c.VisitUserValues(func(key []byte, value any) {
		k := utils.SafeByte2String(key)
		var v string
		switch val := value.(type) {
		case string:
			v = val
		case []byte:
			v = utils.SafeByte2String(val)
		default:
			v = fmt.Sprintf("%v", val)
		}

		params[k] = append(params[k], v)
	})
	return params
}

type EntityReader struct {
	Reader      func(entity proto.Message) error
	SkipMethods map[string]struct{}
}

func (c *Context) DefaultEntityReaders() []*EntityReader {
	return []*EntityReader{
		{Reader: c.ReadQueryParamsToEntity},
		{Reader: c.ReadHeaderParamsToEntity},
		{
			Reader: c.ReadBodyParamsToEntity,
			SkipMethods: map[string]struct{}{
				http.MethodDelete:  struct{}{},
				http.MethodGet:     struct{}{},
				http.MethodConnect: struct{}{},
				http.MethodOptions: struct{}{},
				http.MethodHead:    struct{}{},
				http.MethodTrace:   struct{}{},
			},
		},
		{Reader: c.ReadPathParamsToEntity},
	}
}

func (c *Context) ReadQueryParamsToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.ReadQueryParams()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read query params to entity fail: %v", err)
	}
	return nil
}

func (c *Context) ReadHeaderParamsToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.ReadHeaderParams()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read header params to entity fail: %v", err)
	}
	return nil
}
func (c *Context) ReadPathParamsToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.ReadPathParams()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read path params to entity fail: %v", err)
	}
	return nil
}

func (c *Context) ReadBodyParamsToEntity(entity proto.Message) error {
	body := c.JSONBodyParams()
	if len(body) > 0 {
		if err := protojson.Unmarshal(body, entity); err != nil {
			return status.Errorf(codes.InvalidArgument, "read body params to entity fail: %v", err)
		}
	}
	return nil
}
