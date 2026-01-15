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
	MIME_XML   = "application/xml"
	MIME_JSON  = "application/json"
	MIME_ZIP   = "application/zip"
	MIME_OCTET = "application/octet-stream"
)

// Context wraps fasthttp.RequestCtx to provide helper methods for parameter
// extraction, data binding, and response writing.
type Context struct {
	*fasthttp.RequestCtx
	errPage string
	write   Writer
}

// contextPool implements Object Pooling to reduce GC overhead by reusing Context objects.
var (
	contextPool = sync.Pool{
		New: func() any {
			return &Context{}
		},
	}
)

// NewContext acquires a Context from the pool and initializes it with the given options.
func NewContext(ctx *fasthttp.RequestCtx, options ...Option) *Context {
	c := contextPool.Get().(*Context)
	c.RequestCtx = ctx
	for _, opt := range options {
		opt(c)
	}
	return c
}

// ReadEntity parses request parameters and serializes them into a Protobuf message.
// The default order is: Query -> Header -> Body -> Path.
// Later parameters override earlier ones if keys collide.
func (c *Context) ReadEntity(entity proto.Message) error {
	if entity == nil {
		return nil
	}
	return c.ReadEntityWithReaders(entity, c.DefaultEntityReaders())
}

// ReadEntityWithReaders executes a specific list of entity readers (sources).
func (c *Context) ReadEntityWithReaders(entity proto.Message, readers []*EntityReader) error {
	requestMethod := string(c.Method())
	for _, source := range readers {
		// Skip specific readers (like Body reader) for methods like GET or DELETE.
		if _, ok := source.SkipMethods[requestMethod]; ok {
			continue
		}
		if err := source.Reader(entity); err != nil {
			return err
		}
	}
	return nil
}

// GetUserParam retrieves parameters stored in the context by middleware (Path/User values).
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

// GetHeaderParam retrieves values for a specific HTTP header key.
func (c *Context) GetHeaderParam(key string) []string {
	v := c.Request.Header.PeekAll(key)
	s := make([]string, len(v))
	for idx, b := range v {
		s[idx] = utils.SafeByte2String(b)
	}
	return s
}

// GetQueryParam retrieves values for a specific URL query parameter.
func (c *Context) GetQueryParam(key string) []string {
	v := c.QueryArgs().PeekMulti(key)
	s := make([]string, len(v))
	for idx, b := range v {
		s[idx] = utils.SafeByte2String(b)
	}
	return s
}

// WriteData finalizes the request by sending data or an error back to the client.
func (c *Context) WriteData(data any, err error) {
	if c.write == nil {
		DefaultWriter(c, data, err)
	} else {
		c.write(c, data, err)
	}
	c.Close()
}

// NewOutgoingContext converts the HTTP headers into gRPC-compatible outgoing metadata.
func (c *Context) NewOutgoingContext() context.Context {
	return metadata.NewOutgoingContext(c, c.ReadHeaderParams())
}

// Close cleans up the context and returns it to the pool for reuse.
func (c *Context) Close() {
	c.write = nil
	c.RequestCtx = nil
	contextPool.Put(c)
}

// JSONBodyParams returns the raw body if the Content-Type is application/json.
func (c *Context) JSONBodyParams() []byte {
	if bytes.Equal(c.Request.Header.ContentType(), []byte(MIME_JSON)) {
		return c.Request.Body()
	}
	return []byte{}
}

// ReadQueryParams extracts all query string arguments.
func (c *Context) ReadQueryParams() map[string][]string {
	queries := make(map[string][]string)
	c.QueryArgs().All()(func(key, value []byte) bool {
		k := utils.SafeByte2String(key)
		queries[k] = append(queries[k], utils.SafeByte2String(value))
		return true
	})
	return queries
}

// ReadHeaderParams extracts all HTTP request headers.
func (c *Context) ReadHeaderParams() map[string][]string {
	headers := make(map[string][]string)
	c.Request.Header.All()(func(key, value []byte) bool {
		k := utils.SafeByte2String(key)
		headers[k] = append(headers[k], utils.SafeByte2String(value))
		return true
	})
	return headers
}

// ReadPathParams extracts all variable path segments (e.g., /user/{id}).
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

// EntityReader defines a function to read data into a message and its applicable constraints.
type EntityReader struct {
	Reader      func(entity proto.Message) error
	SkipMethods map[string]struct{}
}

// DefaultEntityReaders defines the standard pipeline for populating a request object.
func (c *Context) DefaultEntityReaders() []*EntityReader {
	return []*EntityReader{
		{Reader: c.ReadQueryParamsToEntity},
		{Reader: c.ReadHeaderParamsToEntity},
		{
			Reader: c.ReadBodyParamsToEntity,
			SkipMethods: map[string]struct{}{
				http.MethodDelete:  {},
				http.MethodGet:     {},
				http.MethodConnect: {},
				http.MethodOptions: {},
				http.MethodHead:    {},
				http.MethodTrace:   {},
			},
		},
		{Reader: c.ReadPathParamsToEntity},
	}
}

// ReadQueryParamsToEntity parses URL queries into the Protobuf struct.
func (c *Context) ReadQueryParamsToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.ReadQueryParams()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read query params to entity fail: %v", err)
	}
	return nil
}

// ReadHeaderParamsToEntity parses HTTP headers into the Protobuf struct.
func (c *Context) ReadHeaderParamsToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.ReadHeaderParams()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read header params to entity fail: %v", err)
	}
	return nil
}

// ReadPathParamsToEntity parses path variables into the Protobuf struct.
func (c *Context) ReadPathParamsToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.ReadPathParams()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read path params to entity fail: %v", err)
	}
	return nil
}

// ReadBodyParamsToEntity unmarshals the JSON body into the Protobuf struct.
func (c *Context) ReadBodyParamsToEntity(entity proto.Message) error {
	body := c.JSONBodyParams()
	if len(body) > 0 {
		if err := protojson.Unmarshal(body, entity); err != nil {
			return status.Errorf(codes.InvalidArgument, "read body params to entity fail: %v", err)
		}
	}
	return nil
}
