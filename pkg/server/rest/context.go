package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/asjard/asjard/core/config"
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
	errorHandler ErrorHandler
	errPage      string
	queryParams  map[string][]string
	// 列表顺序为倒序
	// 例如路由为: /region/{region_id}/project/{project_id}/user/{user_id}
	// 请求路径为: /region/1/project/2/user/3
	// 则解析道此列表中的参数为
	// [{key: "user_id", value: "3"},{key: "project_id", value:"2"},{key: "region_id",value:"1"}]
	pathParams   []*KV
	headerParams map[string][]string
	postBody     []byte
}

// KV .
type KV struct {
	Key, Value string
}

var contextPool = sync.Pool{
	New: func() any {
		return &Context{
			errPage:      config.GetString("servers.rest.doc.errPage", ""),
			queryParams:  make(map[string][]string),
			headerParams: make(map[string][]string),
		}
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
	if err := c.readQueryParamToEntity(entityPtr); err != nil {
		return err
	}
	if err := c.readHeaderParamsToEntity(entityPtr); err != nil {
		return err
	}
	if err := c.readBodyParamsToEntity(entityPtr); err != nil {
		return err
	}
	return c.readPathParamsToEntity(entityPtr)
}

// readQueryEntity 功能同readEntity但不解析body体
// 解析顺序 query -> header -> path
func (c *Context) readQueryEntity(entityPtr any) error {
	if err := c.readQueryParamToEntity(entityPtr); err != nil {
		return err
	}
	if err := c.readHeaderParamsToEntity(entityPtr); err != nil {
		return err
	}
	return c.readPathParamsToEntity(entityPtr)
}

// Write .
func (c *Context) Write(response any, err error) {
	if err != nil {
		c.errorHandler(c, err)
		return
	}
	c.writeJSON(http.StatusOK, &Response{
		Status: &Status{},
		Data:   response,
	})
	c.Close()
}

// Close .
func (c *Context) Close() {
	c.queryParams = make(map[string][]string)
	c.headerParams = make(map[string][]string)
	c.pathParams = []*KV{}
	c.postBody = []byte{}
	contextPool.Put(c)
}

func (c *Context) readBodyParamsToEntity(entity any) error {
	c.postBody = c.Request.Body()
	if entity == nil {
		return nil
	}
	if err := json.Unmarshal(c.postBody, entity); err != nil {
		// 修改下原本返回的错误信息，去掉语言相关内容
		if e, ok := err.(*json.UnmarshalTypeError); ok {
			if e.Struct != "" || e.Field != "" {
				return Error(http.StatusBadRequest,
					"cannot deserialize "+e.Value+" into field "+e.Field+" of type "+e.Type.String(),
					"")
			}
			return Error(http.StatusBadRequest,
				"cannot deserialize "+e.Value+" into value of type "+e.Type.String(),
				"")
		}
		return Error(http.StatusBadRequest, fmt.Sprintf("read body params fail: %s", err.Error()), "")
	}
	return nil
}

func (c *Context) readQueryParamToEntity(entity any) error {
	c.QueryArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)
		if _, ok := c.queryParams[k]; !ok {
			c.queryParams[k] = []string{v}
		} else {
			c.queryParams[k] = append(c.queryParams[k], v)
		}
	})
	if entity == nil {
		return nil
	}
	if err := mapForm(entity, c.queryParams); err != nil {
		return Error(http.StatusBadRequest, fmt.Sprintf("read query params fail: %s", err.Error()), "")
	}
	return nil
}

func (c *Context) readHeaderParamsToEntity(entity any) error {
	c.Request.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)
		if _, ok := c.headerParams[k]; !ok {
			c.headerParams[k] = []string{v}
		} else {
			c.headerParams[k] = append(c.headerParams[k], v)
		}
	})
	if entity == nil {
		return nil
	}
	if err := mapForm(entity, c.headerParams); err != nil {
		return Error(http.StatusBadRequest, fmt.Sprintf("read header params fail: %s", err.Error()), "")
	}
	return nil
}

func (c *Context) readPathParamsToEntity(entity any) error {
	c.VisitUserValues(func(key []byte, value any) {
		c.pathParams = append(c.pathParams, &KV{
			Key:   string(key),
			Value: cast.ToString(value),
		})
	})
	if entity == nil {
		return nil
	}
	pathForm := make(map[string][]string)
	for _, kv := range c.pathParams {
		pathForm[kv.Key] = []string{kv.Value}
	}
	if err := mapForm(entity, pathForm); err != nil {
		return Error(http.StatusBadRequest, fmt.Sprintf("read path params fail: %s", err.Error()), "")
	}
	return nil
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
