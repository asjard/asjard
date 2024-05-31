package server

import (
	"context"
	"sync"
)

// Context .
type Context struct {
	parent   context.Context
	handlers *ServerHandlers
}

var contextPool = sync.Pool{
	New: func() any {
		return &Context{}
	},
}

// NewContext .
func NewContext(parent context.Context, handlers *ServerHandlers) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.parent = parent
	ctx.handlers = handlers
	return ctx
}

// GetParent 获取父上下文
func (c *Context) GetParent() context.Context {
	return c.parent
}

// Handle .
func (c *Context) Handle(handler func()) {
	if c.handlers != nil {
		c.handlers.BeforeRequestHandle(c)
	}
	handler()
	if c.handlers != nil {
		c.handlers.AfterRequestHandle(c)
	}
	c.close()
}

// Close .
func (c *Context) close() {
	c.parent = nil
	c.handlers = nil
	contextPool.Put(c)
}
