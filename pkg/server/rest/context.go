package rest

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/valyala/fasthttp"
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
	errorHandler ErrorHandler
	errPage      string
}

var contextPool = sync.Pool{
	New: func() any {
		logger.Debugf("new context")
		return &Context{}
	},
}

// NewContext .
func NewContext(ctx *fasthttp.RequestCtx) *Context {
	c := contextPool.Get().(*Context)
	c.RequestCtx = ctx
	c.errorHandler = errorHandler
	c.errPage = config.GetString("servers.rest.doc.errPage", "")
	return c
}

// ReadEntity 解析请求参数并序列化到entityPrt中
func (c *Context) ReadEntity(entityPtr any) error {
	switch string(c.Method()) {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return c.readEntity(entityPtr)
	default:
		return c.readQueryEntity(entityPtr)
	}
}

// readEntity 解析请求参数并序列化到entityPrt中
// 解析顺序 query -> header -> body -> path
// 后解析的同名参数会覆盖前解析的同名参数
func (c *Context) readEntity(entityPtr any) error {
	return nil
}

// readQueryEntity 功能同readEntity但不解析body体
// 解析顺序 query -> header -> path
func (c *Context) readQueryEntity(entityPtr any) error {
	return nil
}

func (c *Context) Write(response any, err error) {
	if err != nil {
		c.errorHandler(c, err)
	}
	c.writeJSON(http.StatusOK, &Response{
		Status: &Status{},
		Data:   response,
	})
	contextPool.Put(c)
}

func (c *Context) writeJSON(statusCode int, body any) error {
	if body == nil {
		c.Response.SetStatusCode(statusCode)
		return nil
	}
	c.Response.Header.Set(fasthttp.HeaderContentType, MIME_JSON)
	c.Response.SetStatusCode(statusCode)
	return json.NewEncoder(c.Response.BodyWriter()).Encode(body)
}
