package mcp

import (
	"context"

	"github.com/asjard/asjard/core/server"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Handler defines the interface for services that provide MCP descriptors.
type Handler interface {
	McpServiceDesc() *ServiceDesc
}

// ToolHandler is the standard signature for an MCP tool.
// It wraps business logic with Asjard unary interceptors.
type ToolHandler func(ctx context.Context, req *mcpsdk.CallToolRequest, srv any, interceptor server.UnaryServerInterceptor) (*mcpsdk.CallToolResult, error)

// ServiceDesc represents an MCP service's specification.
type ServiceDesc struct {
	ServiceName string     // Unique identifier for the service (e.g., "api.v1.user").
	Name        string     // Human-readable display name.
	Desc        string     // Brief description of the service.
	HandlerType any        // Pointer to the interface type, used for reflection validation.
	Tools       []ToolDesc // Tools exposed by this service.
}

// ToolDesc represents a single MCP tool.
type ToolDesc struct {
	Name         string
	Title        string
	Description  string
	InputSchema  any
	OutputSchema any
	Handler      ToolHandler
}
