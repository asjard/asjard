package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/pkg/status"
	"github.com/spf13/cast"
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
	// errorHandler ErrorHandler
	errPage     string
	queryParams map[string][]string
	queryLoaded bool
	// 例如路由为: /region/{region_id}/project/{project_id}/user/{user_id}
	// 请求路径为: /region/1/project/2/user/3
	// {"region_id":"1","project_id":"2","user_id":"3"}
	pathParams   map[string][]string
	pathLoaded   bool
	headerParams map[string][]string
	headLoaded   bool
	postBody     []byte
	postLoaded   bool
	// 返回内容
	response *Response
	write    Writer
}

// KV .
// type KV struct {
// 	Key, Value string
// }

var contextPool = sync.Pool{
	New: func() any {
		return &Context{
			errPage:      config.GetString("servers.rest.doc.errPage", ""),
			queryParams:  make(map[string][]string),
			pathParams:   make(map[string][]string),
			headerParams: make(map[string][]string),
		}
	},
}

// NewContext .
func NewContext(ctx *fasthttp.RequestCtx, options ...Option) *Context {
	c := contextPool.Get().(*Context)
	c.RequestCtx = ctx
	c.errPage = config.GetString("servers.rest.doc.errPage", "")
	c.write = defaultWriter
	for _, opt := range options {
		opt(c)
	}
	return c
}

// ReadEntity 解析请求参数并序列化到entityPrt中
// 解析顺序 query -> header -> body -> path
// 后解析的同名参数会覆盖前解析的同名参数
// post,put,patch解析body体,其余不解析
func (c *Context) ReadEntity(entityPtr any) error {
	if err := c.readQueryParamsToEntity(entityPtr); err != nil {
		return err
	}
	if err := c.readHeaderParamsToEntity(entityPtr); err != nil {
		return err
	}
	requestMethod := string(c.Method())
	if requestMethod == http.MethodPost ||
		requestMethod == http.MethodPut ||
		requestMethod == http.MethodPatch {
		if err := c.readBodyParamsToEntity(entityPtr); err != nil {
			return err
		}
	}
	return c.readPathParamsToEntity(entityPtr)
}

// ReadAndWrite 解析请求参数并返回
func (c *Context) ReadAndWrite(handler func(ctx *Context, in any) (any, error), entityPtr any) {
	if err := c.ReadEntity(entityPtr); err != nil {
		c.Write(nil, err)
		return
	}
	c.Write(handler(c, entityPtr))
}

// GetParam 获取参数
// 获取顺序 path -> header -> query
// 返回获取到的第一个
func (c *Context) GetParam(key string) (string, bool) {
	if value, ok := c.GetPathParam(key); ok {
		return value, ok
	}
	if value, ok := c.GetHeaderParam(key); ok {
		return value, ok
	}
	return c.GetQueryParam(key)
}

// GetPathParam 获取路径参数
func (c *Context) GetPathParam(key string) (string, bool) {
	c.readPathParams()
	values, ok := c.pathParams[key]
	if !ok {
		return "", false
	}
	if len(values) == 0 {
		return "", true
	}
	return values[0], true
}

// GetHeaderParam 获取请求头参数
func (c *Context) GetHeaderParam(key string) (string, bool) {
	c.readHeaderParams()
	values, ok := c.headerParams[key]
	if !ok {
		return "", false
	}
	if len(values) == 0 {
		return "", true
	}
	return values[0], true
}

// GetQueryParam 获取query参数
func (c *Context) GetQueryParam(key string) (string, bool) {
	c.readQueryParams()
	values, ok := c.queryParams[key]
	if !ok {
		return "", false
	}
	if len(values) == 0 {
		return "", true
	}
	return values[0], true
}

// Write 请求返回
func (c *Context) Write(data any, err error) {
	c.write(c, data, err)
	c.Close()
}

// Close .
func (c *Context) Close() {
	c.queryParams = make(map[string][]string)
	c.queryLoaded = false
	c.headerParams = make(map[string][]string)
	c.headLoaded = false
	c.pathParams = make(map[string][]string)
	c.pathLoaded = false
	c.postBody = []byte{}
	c.postLoaded = false
	contextPool.Put(c)
}

func (c *Context) readBodyParams() {
	if !c.postLoaded {
		c.postBody = c.Request.Body()
		c.postLoaded = true
	}
}

func (c *Context) readBodyParamsToEntity(entity any) error {
	c.readBodyParams()
	if entity == nil {
		return nil
	}
	if len(c.postBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(c.postBody, entity); err != nil {
		// 修改下原本返回的错误信息，去掉语言相关内容
		if e, ok := err.(*json.UnmarshalTypeError); ok {
			if e.Struct != "" || e.Field != "" {
				return status.Errorf(http.StatusBadRequest,
					"cannot deserialize "+e.Value+" into field "+e.Field+" of type "+e.Type.String())
			}
			return status.Errorf(http.StatusBadRequest,
				"cannot deserialize "+e.Value+" into value of type "+e.Type.String())
		}
		return status.Errorf(http.StatusBadRequest, fmt.Sprintf("read body params fail: %s", err.Error()))
	}
	return nil
}

func (c *Context) readQueryParams() {
	if !c.queryLoaded {
		c.QueryArgs().VisitAll(func(key, value []byte) {
			k := string(key)
			v := string(value)
			if _, ok := c.queryParams[k]; !ok {
				c.queryParams[k] = []string{v}
			} else {
				c.queryParams[k] = append(c.queryParams[k], v)
			}
		})
		c.queryLoaded = true
	}
}

func (c *Context) readQueryParamsToEntity(entity any) error {
	c.readQueryParams()
	if entity == nil {
		return nil
	}
	if err := mapForm(entity, c.queryParams); err != nil {
		return status.Errorf(http.StatusBadRequest, fmt.Sprintf("read query params fail: %s", err.Error()))
	}
	return nil
}

func (c *Context) readHeaderParams() {
	if !c.headLoaded {
		c.Request.Header.VisitAll(func(key, value []byte) {
			k := string(key)
			v := string(value)
			if _, ok := c.headerParams[k]; !ok {
				c.headerParams[k] = []string{v}
			} else {
				c.headerParams[k] = append(c.headerParams[k], v)
			}
		})
		c.headLoaded = true
	}
}

func (c *Context) readHeaderParamsToEntity(entity any) error {
	c.readBodyParams()
	if entity == nil {
		return nil
	}
	if err := mapForm(entity, c.headerParams); err != nil {
		return status.Errorf(http.StatusBadRequest, fmt.Sprintf("read header params fail: %s", err.Error()))
	}
	return nil
}

func (c *Context) readPathParams() {
	if !c.pathLoaded {
		c.VisitUserValues(func(key []byte, value any) {
			keyStr := string(key)
			valueStr := cast.ToString(value)
			if _, ok := c.pathParams[keyStr]; ok {
				c.pathParams[keyStr] = append(c.pathParams[keyStr], valueStr)
			} else {
				c.pathParams[keyStr] = []string{valueStr}
			}
		})
		c.pathLoaded = true
	}
}

func (c *Context) readPathParamsToEntity(entity any) error {
	c.readPathParams()
	if entity == nil {
		return nil
	}
	if err := mapForm(entity, c.pathParams); err != nil {
		return status.Errorf(http.StatusBadRequest, fmt.Sprintf("read path params fail: %s", err.Error()))
	}
	return nil
}
