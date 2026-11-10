package rest

import (
	"context"
	"io"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
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
	headerXForwardedFor = "X-Forwarded-For"
	headerXRealIP       = "X-Real-IP"
	comma               = ","
)

// Context wraps fasthttp.RequestCtx to provide helper methods for parameter
// extraction, data binding, and response writing.
type Context struct {
	r *http.Request
	w http.ResponseWriter

	// This mutex protects userValues map.
	um         sync.RWMutex
	userValues map[any]any

	errPage string
	writer  Writer
}

// contextPool implements Object Pooling to reduce GC overhead by reusing Context objects.
var (
	contextPool = sync.Pool{
		New: func() any {
			return &Context{
				userValues: make(map[any]any),
			}
		},
	}
)

// NewContext acquires a Context from the pool and initializes it with the given options.
func NewContext(r *http.Request, w http.ResponseWriter, options ...Option) *Context {
	c := contextPool.Get().(*Context)
	c.r = r
	c.w = w
	for _, opt := range options {
		opt(c)
	}
	return c
}

func (c *Context) Request() *http.Request {
	return c.r
}

func (c *Context) Response() http.ResponseWriter {
	return c.w
}

func (c *Context) Context() context.Context {
	return c.r.Context()
}

func (c *Context) WithContext(ctx context.Context) {
	c.r = c.r.WithContext(ctx)
}

func (c *Context) SetUserValue(key, val any) {
	c.um.Lock()
	c.userValues[key] = val
	c.um.Unlock()
}

func (c *Context) GetUserValue(key any) any {
	c.um.RLock()
	v, ok := c.userValues[key]
	c.um.RUnlock()
	if ok {
		return v
	}
	return nil
}

func (c *Context) Deadline() (time.Time, bool) {
	return c.Context().Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Context().Done()
}

func (c *Context) Err() error {
	return c.Context().Err()
}

func (c *Context) Value(key any) any {
	return c.GetUserValue(key)
}

func (c *Context) GetUserValues() map[any]any {
	vals := make(map[any]any)
	c.um.RLock()
	maps.Copy(vals, c.userValues)
	c.um.RUnlock()
	return vals
}

func (c *Context) Query(key string) string {
	return c.r.URL.Query().Get(key)
}
func (c *Context) Queries() map[string][]string {
	res := make(map[string][]string)
	maps.Copy(res, c.r.URL.Query())
	return res
}
func (c *Context) Header(key string) string {
	return c.r.Header.Get(key)
}
func (c *Context) Headers() map[string][]string {
	res := make(map[string][]string)
	maps.Copy(res, c.r.Header)
	return res
}

func (c *Context) PathValue(key string) string {
	return c.r.PathValue(key)
}

func (c *Context) PathValues() map[string][]string {
	res := make(map[string][]string)
	for _, param := range pathParamNames(c.r.URL.Path) {
		val := c.PathValue(param)
		if _, ok := res[param]; ok {
			res[param] = append(res[param], val)
		} else {
			res[param] = []string{val}
		}
	}
	return res
}

func (c *Context) SendFile(filePath string) {
	http.ServeFile(c.w, c.r, filePath)
}

// pathParamNames extracts wildcard names from a ServeMux pattern.
func pathParamNames(pattern string) []string {
	var names []string
	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '{' {
			continue
		}
		j := i + 1
		for j < len(pattern) && pattern[j] != '}' {
			j++
		}
		if j >= len(pattern) {
			break
		}
		name := strings.TrimSuffix(pattern[i+1:j], "...")
		if name != "" {
			names = append(names, name)
		}
		i = j
	}
	return names
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
	for _, source := range readers {
		// Skip specific readers (like Body reader) for methods like GET or DELETE.
		if _, ok := source.SkipMethods[c.r.Method]; ok {
			continue
		}
		if err := source.Reader(entity); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) ClientIP() string {
	xff := c.Header(headerXForwardedFor)
	if len(xff) > 0 {
		return xff
	}

	xri := c.Header(headerXRealIP)
	if len(xri) > 0 {
		return xri
	}
	return c.r.RemoteAddr
}

// WriteData finalizes the request by sending data or an error back to the client.
func (c *Context) WriteData(data any, err error) {
	if c.writer == nil {
		DefaultWriter(c, data, err)
	} else {
		c.writer(c, data, err)
	}
	c.Close()
}

// Write writes the data to the connection as part of an HTTP reply.
func (c *Context) Write(data []byte) (int, error) {
	return c.w.Write(data)
}

func (c *Context) SetWriter(writer Writer) {
	c.writer = writer
}

// Close cleans up the context and returns it to the pool for reuse.
func (c *Context) Close() {
	c.writer = nil
	c.r = nil
	c.w = nil
	c.userValues = nil
	contextPool.Put(c)
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
			Reader: c.ReadJsonBodyToEntity,
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
	if err := protoForm(entity, c.Queries()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read query params to entity fail: %v", err)
	}
	return nil
}

// ReadHeaderParamsToEntity parses HTTP headers into the Protobuf struct.
func (c *Context) ReadHeaderParamsToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.Headers()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read header params to entity fail: %v", err)
	}
	return nil
}

// ReadPathParamsToEntity parses path variables into the Protobuf struct.
func (c *Context) ReadPathParamsToEntity(entity proto.Message) error {
	if err := protoForm(entity, c.PathValues()); err != nil {
		return status.Errorf(codes.InvalidArgument, "read path params to entity fail: %v", err)
	}
	return nil
}

// ReadJsonBodyToEntity unmarshals the JSON body into the Protobuf struct.
func (c *Context) ReadJsonBodyToEntity(entity proto.Message) error {
	if ok := strings.Contains(c.Header("Content-Type"), MIME_JSON); !ok {
		return nil
	}
	b, err := io.ReadAll(c.r.Body)
	if err != nil {
		logger.L(c.Context()).Error("read body fail", "err")
	}
	if len(b) == 0 {
		return nil
	}
	if err := protojson.Unmarshal(b, entity); err != nil {
		return status.Errorf(codes.InvalidArgument, "read body params to entity fail: %v", err)
	}
	return nil
}
