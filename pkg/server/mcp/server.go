package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest/xhttp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// Protocol is the identifier for the MCP server implementation.
	Protocol = "mcp"
)

var (
	emptyInputSchema = json.RawMessage(`{"type":"object"}`)

	_ server.Server = &McpServer{}
)

func init() {
	server.AddServer(Protocol, New)
}

// McpServer is an independent MCP protocol server that reuses xhttp for listening.
type McpServer struct {
	conf        Config
	interceptor server.UnaryServerInterceptor
	mcpServer   *mcpsdk.Server
	httpServer  *xhttp.HttpServer
}

// New initializes the MCP server by loading configurations and setting up certificates.
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig()
	if err := config.GetWithUnmarshal(constant.ConfigServerMcpPrefix, &conf); err != nil {
		return nil, err
	}
	return MustNew(conf, options)
}

// MustNew performs the low-level setup of the MCP server on top of xhttp.
func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	if options == nil {
		options = &server.ServerOptions{}
	}
	hs, err := xhttp.MustNew(conf.Config, options)
	if err != nil {
		return nil, err
	}
	app := runtime.GetAPP()
	return &McpServer{
		conf:        conf,
		interceptor: options.Interceptor,
		httpServer:  hs.(*xhttp.HttpServer),
		mcpServer: mcpsdk.NewServer(&mcpsdk.Implementation{
			Name:    app.Instance.Name,
			Version: app.Instance.Version,
		}, nil),
	}, nil
}

// AddHandler registers MCP tools from a service descriptor.
func (s *McpServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invalid handler %T, must implement *mcp.McpServiceDesc", handler)
	}
	return s.addTools(h)
}

// Start mounts the MCP handler on xhttp and starts the HTTP listener.
func (s *McpServer) Start(startErr chan error) error {
	if s.httpServer == nil {
		return errors.New("mcp http server not initialized")
	}
	opts := &mcpsdk.StreamableHTTPOptions{}
	if s.conf.Options.MaxRequestBodySize > 0 {
		opts.MaxRequestBodyBytes = int64(s.conf.Options.MaxRequestBodySize)
	}
	s.httpServer.Handle(s.conf.Path, mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
		return s.mcpServer
	}, opts))
	if err := s.httpServer.Start(startErr); err != nil {
		return err
	}
	logger.Debug("start mcp server", "address", s.conf.Addresses.Listen, "path", s.conf.Path)
	return nil
}

// Stop performs a graceful shutdown of the underlying HTTP server.
func (s *McpServer) Stop() {
	if s.httpServer != nil {
		s.httpServer.Stop()
	}
}

// Protocol returns the server protocol name.
func (s *McpServer) Protocol() string {
	return Protocol
}

// Enabled checks if the MCP server is enabled in config.
func (s *McpServer) Enabled() bool {
	return s.conf.Enabled
}

// ListenAddresses returns the configured listening addresses.
func (s *McpServer) ListenAddresses() server.AddressConfig {
	return s.conf.Addresses
}

// Register adds tools from an MCP handler to the SDK server.
func (s *McpServer) addTools(handler Handler) error {
	desc := handler.McpServiceDesc()
	if desc == nil {
		return nil
	}
	ht := reflect.TypeOf(desc.HandlerType).Elem()
	st := reflect.TypeOf(handler)
	if !st.Implements(ht) {
		return fmt.Errorf("found the handler of type %v that does not satisfy %v", st, ht)
	}
	for _, tool := range desc.Tools {
		if tool.Name == "" || tool.Handler == nil {
			continue
		}
		s.addTool(tool, handler)
	}
	return nil
}

func (s *McpServer) addTool(tool ToolDesc, svc Handler) {
	schema := tool.InputSchema
	if schema == nil {
		schema = emptyInputSchema
	}
	s.mcpServer.AddTool(&mcpsdk.Tool{
		Name:         tool.Name,
		Title:        tool.Title,
		Description:  tool.Description,
		InputSchema:  schema,
		OutputSchema: tool.OutputSchema,
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		return tool.Handler(ctx, req, svc, s.interceptor)
	})
}
