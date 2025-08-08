package rest

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/status"
	"github.com/spf13/cast"
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

var contextPool = sync.Pool{
	New: func() any {
		return &Context{
			errPage: config.GetString("asjard.service.website", ""),
		}
	},
}

// NewContext .
func NewContext(ctx *fasthttp.RequestCtx, options ...Option) *Context {
	c := &Context{
		RequestCtx: ctx,
	}
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
func (c *Context) GetPathParam(key string) []string {
	return c.ReadPathParams()[strings.ToLower(key)]
}

// GetHeaderParam 获取请求头参数
func (c *Context) GetHeaderParam(key string) []string {
	return c.ReadHeaderParams()[strings.ToLower(key)]
}

// GetQueryParam 获取query参数
func (c *Context) GetQueryParam(key string) []string {
	return c.ReadQueryParams()[strings.ToLower(key)]
}

// WriteData 请求返回
func (c *Context) WriteData(data any, err error) {
	if c.write == nil {
		DefaultWriter(c, data, err)
	} else {
		c.write(c, data, err)
	}
	// c.Close()
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
	contextPool.Put(c)
}

// JSONBodyParams 读取请求体
func (c *Context) JSONBodyParams() []byte {
	if string(c.Request.Header.ContentType()) == MIME_JSON {
		return c.Request.Body()
	}
	return []byte{}
}

// ReadQueryParams 获取query参数
func (c *Context) ReadQueryParams() map[string][]string {
	params := make(map[string][]string)
	c.QueryArgs().All()(func(key, value []byte) bool {
		k := string(key)
		v := string(value)
		if _, ok := params[k]; !ok {
			params[k] = []string{v}
		} else {
			params[k] = append(params[k], v)
		}
		return true
	})
	return params
}

// ReadHeaderParams 读取header请求参数
func (c *Context) ReadHeaderParams() map[string][]string {
	params := make(map[string][]string)
	c.Request.Header.All()(func(key, value []byte) bool {
		k := strings.ToLower(string(key))
		v := string(value)
		if _, ok := params[k]; !ok {
			params[k] = []string{v}
		} else {
			params[k] = append(params[k], v)
		}
		return true
	})
	return params
}

// ReadPathParams 读取path请求参数
func (c *Context) ReadPathParams() map[string][]string {
	params := make(map[string][]string)
	c.VisitUserValues(func(key []byte, value any) {
		keyStr := string(key)
		valueStr := cast.ToString(value)
		if _, ok := params[keyStr]; ok {
			params[keyStr] = append(params[keyStr], valueStr)
		} else {
			params[keyStr] = []string{valueStr}
		}
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
	if err := protojson.Unmarshal(c.JSONBodyParams(), entity); err != nil {
		return status.Errorf(codes.InvalidArgument, "read body params to entity fail: %v", err)
	}
	return nil
}
