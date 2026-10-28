package xfasthttp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	MIME_XML   = "application/xml"
	MIME_JSON  = "application/json"
	MIME_ZIP   = "application/zip"
	MIME_OCTET = "application/octet-stream"
)

var (
	_ rest.RestContext = (*Context)(nil)

	headerXForwardedFor = []byte("X-Forwarded-For")
	headerXRealIP       = []byte("X-Real-IP")
	comma               = []byte(",")
)

// Context is the fasthttp implementation of rest.RestContext.
type Context struct {
	*fasthttp.RequestCtx
	errPage     string
	write       Writer
	i18nEnabled bool
}

type ctxKey int32

const (
	userContext ctxKey = 1
)

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

func (c *Context) Context() context.Context {
	if ctx, ok := c.UserValue(userContext).(context.Context); ok && ctx != nil {
		return ctx
	}
	return context.Background()
}

func (c *Context) SetContext(ctx context.Context) {
	c.SetUserValue(userContext, ctx)
}

// IsI18nEnabled return i18n is enabled
func (c *Context) IsI18nEnabled() bool {
	return c.i18nEnabled
}

// Method return request method
func (c *Context) Method() string {
	return utils.SafeByte2String(c.RequestCtx.Method())
}

// Path return request path
func (c *Context) Path() string {
	return utils.SafeByte2String(c.RequestCtx.Path())
}

// RequestSize return request size
func (c *Context) RequestSize() int {
	if c.RequestCtx == nil {
		return 0
	}
	s := len(c.RequestURI())
	s += len(c.Method())
	s += len(c.Request.Header.Protocol())
	s += c.Request.Header.Len()
	s += c.Request.Header.ContentLength()
	return s
}

// ResponseSize return response size
func (c *Context) ResponseSize() int {
	if c.RequestCtx == nil {
		return 0
	}
	n := c.Response.Header.ContentLength()
	if n < 0 {
		return 0
	}
	return n
}

// UserValue returns the value stored via SetUserValue* under the given key.
func (c *Context) UserValue(key any) any {
	return c.RequestCtx.UserValue(key)
}

// SetUserValue stores the given value (arbitrary object)
// under the given key in Request.
func (c *Context) SetUserValue(key, value any) {
	c.RequestCtx.SetUserValue(key, value)
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

// Header retrieves value for a specific HTTP header key.
func (c *Context) Header(key string) string {
	v := c.Request.Header.Peek(key)
	if len(v) == 0 {
		return ""
	}
	return utils.SafeByte2String(v)
}

// ClientIP returns the client ip the request came from.
func (c *Context) ClientIP() string {
	xff := c.Request.Header.PeekBytes(headerXForwardedFor)
	if len(xff) > 0 {
		before, _, _ := bytes.Cut(xff, []byte(","))
		return utils.SafeByte2String(bytes.TrimSpace(before))
	}

	xri := c.Request.Header.PeekBytes(headerXRealIP)
	if len(xri) > 0 {
		return utils.SafeByte2String(bytes.TrimSpace(xri))
	}

	return c.RemoteIP().String()
}

// Query retrieves value for a specific URL query parameter.
func (c *Context) Query(key string) string {
	v := c.QueryArgs().Peek(key)
	if len(v) == 0 {
		return ""
	}
	return utils.SafeByte2String(v)
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

// SetWriter
func (c *Context) SetWriter(writer Writer) {
	c.write = writer
}

// Close cleans up the context and returns it to the pool for reuse.
func (c *Context) Close() {
	c.write = nil
	c.RequestCtx = nil
	c.errPage = ""
	c.i18nEnabled = false
	contextPool.Put(c)
}

// JSONBody returns the raw body if the Content-Type is application/json.
func (c *Context) JSONBody() []byte {
	if bytes.Contains(c.Request.Header.ContentType(), []byte(MIME_JSON)) {
		return c.Request.Body()
	}
	logger.L(c).Warn("request contentType not json", "content_type", utils.SafeByte2String(c.Request.Header.ContentType()))
	return []byte{}
}

// Queries extracts all query string arguments.
func (c *Context) Queries() map[string][]string {
	queries := make(map[string][]string)
	c.QueryArgs().All()(func(key, value []byte) bool {
		k := utils.SafeByte2String(key)
		queries[k] = append(queries[k], utils.SafeByte2String(value))
		return true
	})
	return queries
}

// Headers extracts all HTTP request headers.
func (c *Context) Headers() map[string][]string {
	headers := make(map[string][]string)
	c.Request.Header.All()(func(key, value []byte) bool {
		k := utils.SafeByte2String(key)
		headers[k] = append(headers[k], utils.SafeByte2String(value))
		return true
	})
	return headers
}

// UserValues extracts all variable path segments (e.g., /user/{id}).
func (c *Context) UserValues() map[string][]string {
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
		{Reader: c.ReadQueryToEntity},
		{Reader: c.ReadHeaderToEntity},
		{
			Reader: c.ReadBodyToEntity,
			SkipMethods: map[string]struct{}{
				http.MethodDelete:  {},
				http.MethodGet:     {},
				http.MethodConnect: {},
				http.MethodOptions: {},
				http.MethodHead:    {},
				http.MethodTrace:   {},
			},
		},
		{Reader: c.ReadPathToEntity},
	}
}

// ReadQueryToEntity parses URL queries into the Protobuf struct.
func (c *Context) ReadQueryToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.Queries()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read query params to entity fail: %v", err)
	}
	return nil
}

// ReadHeaderToEntity parses HTTP headers into the Protobuf struct.
func (c *Context) ReadHeaderToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.Headers()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read header params to entity fail: %v", err)
	}
	return nil
}

// ReadPathToEntity parses path variables into the Protobuf struct.
func (c *Context) ReadPathToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.UserValues()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read path params to entity fail: %v", err)
	}
	return nil
}

// ReadBodyToEntity unmarshals the JSON body into the Protobuf struct.
func (c *Context) ReadBodyToEntity(entity proto.Message) error {
	body := c.JSONBody()
	if len(body) > 0 {
		if err := protojson.Unmarshal(body, entity); err != nil {
			return status.Errorf(codes.InvalidArgument, "read body params to entity fail: %v", err)
		}
	} else {
		logger.L(c).Warn("read body params is empty", "request_method", c.Method())
	}
	return nil
}
