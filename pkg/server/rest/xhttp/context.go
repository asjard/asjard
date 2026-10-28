package xhttp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	MIME_XML   = "application/xml"
	MIME_JSON  = "application/json"
	MIME_ZIP   = "application/zip"
	MIME_OCTET = "application/octet-stream"

	headerContentType                   = "Content-Type"
	headerAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	headerAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	headerAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	headerAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	headerAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	headerAccessControlMaxAge           = "Access-Control-Max-Age"
	headerAccessControlRequestMethod    = "Access-Control-Request-Method"
	headerAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	headerXForwardedFor                 = "X-Forwarded-For"
	headerXRealIP                       = "X-Real-IP"
)

// Context is the net/http implementation of rest.RestContext.
type Context struct {
	*RequestCtx
	errPage     string
	write       Writer
	i18nEnabled bool
}

// RequestCtx wraps net/http request and response objects, mirroring fasthttp.RequestCtx.
type RequestCtx struct {
	Request  *http.Request
	Response *Response
	values   map[any]any
	body     []byte
	bodyRead bool
	bodyErr  error
}

// Response tracks status, size, and the underlying http.ResponseWriter.
type Response struct {
	writer      http.ResponseWriter
	statusCode  int
	size        int
	wroteHeader bool
}

type ctxKey int32

const (
	userContext ctxKey = 1
)

var (
	_ context.Context     = (*RequestCtx)(nil)
	_ context.Context     = (*Context)(nil)
	_ rest.RestContext    = (*Context)(nil)
	_ http.ResponseWriter = (*Response)(nil)

	contextPool = sync.Pool{
		New: func() any {
			return &Context{}
		},
	}
)

// NewRequestCtx wraps a net/http request pair as a RequestCtx.
func NewRequestCtx(w http.ResponseWriter, r *http.Request) *RequestCtx {
	ctx := &RequestCtx{
		Request: r,
		Response: &Response{
			writer:     w,
			statusCode: http.StatusOK,
		},
		values: make(map[any]any),
	}
	if r != nil {
		for _, name := range pathParamNames(r.Pattern) {
			if v := r.PathValue(name); v != "" {
				ctx.values[name] = v
			}
		}
	}
	return ctx
}

// NewContext acquires a Context from the pool and initializes it with the given options.
func NewContext(ctx *RequestCtx, options ...Option) *Context {
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
	if c.Request == nil {
		return ""
	}
	return c.Request.Method
}

// Path return request path
func (c *Context) Path() string {
	if c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.Path
}

// RequestSize return request size
func (c *Context) RequestSize() int {
	if c.RequestCtx == nil || c.Request == nil {
		return 0
	}
	s := len(c.RequestURI())
	s += len(c.Method())
	s += len(c.Request.Proto)
	for k, vs := range c.Request.Header {
		s += len(k)
		for _, v := range vs {
			s += len(v)
		}
	}
	if n := int(c.Request.ContentLength); n > 0 {
		s += n
	}
	return s
}

// ResponseSize return response size
func (c *Context) ResponseSize() int {
	if c.RequestCtx == nil || c.Response == nil {
		return 0
	}
	return c.Response.size
}

// UserValue returns the value stored via SetUserValue under the given key.
func (c *Context) UserValue(key any) any {
	if c.RequestCtx == nil {
		return nil
	}
	return c.RequestCtx.UserValue(key)
}

// SetUserValue stores the given value (arbitrary object)
// under the given key in Request.
func (c *Context) SetUserValue(key, value any) {
	if c.RequestCtx == nil {
		return
	}
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
	requestMethod := c.Method()
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

// Header retrieves value for a specific HTTP header key.
func (c *Context) Header(key string) string {
	if c.Request == nil {
		return ""
	}
	return c.Request.Header.Get(key)
}

// ClientIP returns the client ip the request came from.
func (c *Context) ClientIP() string {
	if c.Request == nil {
		return ""
	}
	if xff := c.Request.Header.Get(headerXForwardedFor); xff != "" {
		before, _, _ := strings.Cut(xff, ",")
		return strings.TrimSpace(before)
	}
	if xri := c.Request.Header.Get(headerXRealIP); xri != "" {
		return strings.TrimSpace(xri)
	}
	if ip := c.RemoteIP(); ip != nil {
		return ip.String()
	}
	return c.Request.RemoteAddr
}

// Query retrieves value for a specific URL query parameter.
func (c *Context) Query(key string) string {
	return c.QueryArgs().Get(key)
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

// SetWriter overrides the response writer for this request.
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
	if c.Request == nil {
		return []byte{}
	}
	contentType := c.Request.Header.Get(headerContentType)
	if !strings.Contains(contentType, MIME_JSON) {
		logger.L(c).Warn("request contentType not json", "content_type", contentType)
		return []byte{}
	}
	return c.readBody()
}

// Queries extracts all query string arguments.
func (c *Context) Queries() map[string][]string {
	args := c.QueryArgs()
	queries := make(map[string][]string, len(args))
	for k, vs := range args {
		queries[k] = append([]string(nil), vs...)
	}
	return queries
}

// Headers extracts all HTTP request headers.
func (c *Context) Headers() map[string][]string {
	if c.Request == nil {
		return map[string][]string{}
	}
	headers := make(map[string][]string, len(c.Request.Header))
	for k, vs := range c.Request.Header {
		headers[k] = append([]string(nil), vs...)
	}
	return headers
}

// UserValues extracts all variable path segments (e.g., /user/{id}).
func (c *Context) UserValues() map[string][]string {
	params := make(map[string][]string)
	if c.RequestCtx == nil {
		return params
	}
	c.VisitUserValues(func(key []byte, value any) {
		k := string(key)
		var v string
		switch val := value.(type) {
		case string:
			v = val
		case []byte:
			v = string(val)
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

func (c *RequestCtx) UserValue(key any) any {
	if c == nil || c.values == nil {
		return nil
	}
	return c.values[key]
}

func (c *RequestCtx) SetUserValue(key, value any) {
	if c == nil {
		return
	}
	if c.values == nil {
		c.values = make(map[any]any)
	}
	c.values[key] = value
}

func (c *RequestCtx) VisitUserValues(f func(key []byte, value any)) {
	if c == nil || c.values == nil {
		return
	}
	for k, v := range c.values {
		var key []byte
		switch kk := k.(type) {
		case string:
			key = []byte(kk)
		case []byte:
			key = kk
		default:
			key = fmt.Appendf(nil, "%v", kk)
		}
		f(key, v)
	}
}

func (c *RequestCtx) QueryArgs() url.Values {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return url.Values{}
	}
	return c.Request.URL.Query()
}

func (c *RequestCtx) RequestURI() string {
	if c == nil || c.Request == nil {
		return ""
	}
	if c.Request.RequestURI != "" {
		return c.Request.RequestURI
	}
	if c.Request.URL != nil {
		return c.Request.URL.RequestURI()
	}
	return ""
}

func (c *RequestCtx) readBody() []byte {
	if c == nil {
		return nil
	}
	if c.bodyRead {
		return c.body
	}
	c.bodyRead = true
	if c.Request == nil || c.Request.Body == nil {
		return nil
	}
	b, err := io.ReadAll(c.Request.Body)
	_ = c.Request.Body.Close()
	c.body = b
	c.bodyErr = err
	if err != nil {
		logger.L(c).Warn("read body fail", "err", err)
		return []byte{}
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(b))
	return b
}

func (c *RequestCtx) RemoteIP() net.IP {
	if c == nil || c.Request == nil {
		return nil
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return net.ParseIP(c.Request.RemoteAddr)
	}
	return net.ParseIP(host)
}

func (c *RequestCtx) Write(b []byte) (int, error) {
	if c == nil || c.Response == nil {
		return 0, http.ErrNotSupported
	}
	return c.Response.Write(b)
}

func (c *RequestCtx) SetStatusCode(code int) {
	if c == nil || c.Response == nil {
		return
	}
	c.Response.SetStatusCode(code)
}

func (c *RequestCtx) SetContentType(contentType string) {
	if c == nil || c.Response == nil {
		return
	}
	c.Response.Header().Set(headerContentType, contentType)
}

func (c *RequestCtx) SendFile(path string) {
	if c == nil || c.Response == nil || c.Request == nil {
		return
	}
	http.ServeFile(c.Response, c.Request, path)
}

func (c *RequestCtx) Redirect(uri string, statusCode int) {
	if c == nil || c.Response == nil || c.Request == nil {
		return
	}
	http.Redirect(c.Response, c.Request, uri, statusCode)
}

func (c *RequestCtx) Deadline() (time.Time, bool) {
	if c == nil || c.Request == nil {
		return time.Time{}, false
	}
	return c.Request.Context().Deadline()
}

func (c *RequestCtx) Done() <-chan struct{} {
	if c == nil || c.Request == nil {
		return nil
	}
	return c.Request.Context().Done()
}

func (c *RequestCtx) Err() error {
	if c == nil || c.Request == nil {
		return nil
	}
	return c.Request.Context().Err()
}

func (c *RequestCtx) Value(key any) any {
	if v := c.UserValue(key); v != nil {
		return v
	}
	if c == nil || c.Request == nil {
		return nil
	}
	return c.Request.Context().Value(key)
}

func (r *Response) Header() http.Header {
	if r == nil || r.writer == nil {
		return http.Header{}
	}
	return r.writer.Header()
}

func (r *Response) SetStatusCode(code int) {
	if r == nil {
		return
	}
	r.statusCode = code
}

func (r *Response) WriteHeader(code int) {
	if r == nil || r.writer == nil || r.wroteHeader {
		return
	}
	r.statusCode = code
	r.wroteHeader = true
	r.writer.WriteHeader(code)
}

func (r *Response) Write(b []byte) (int, error) {
	if r == nil || r.writer == nil {
		return 0, http.ErrNotSupported
	}
	if !r.wroteHeader {
		r.WriteHeader(r.statusCode)
	}
	n, err := r.writer.Write(b)
	r.size += n
	return n, err
}

func (r *Response) Unwrap() http.ResponseWriter {
	if r == nil {
		return nil
	}
	return r.writer
}

func (r *Response) Flush() {
	if r == nil {
		return
	}
	if f, ok := r.writer.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if r == nil {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	h, ok := r.writer.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return h.Hijack()
}
