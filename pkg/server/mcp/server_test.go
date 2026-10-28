package mcp

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/server"
	_ "github.com/asjard/asjard/pkg/config/mem"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type greetServer interface {
	Greet(ctx context.Context, name string) (string, error)
}

type greetAPI struct{}

func (g greetAPI) Greet(_ context.Context, name string) (string, error) {
	if name == "" {
		name = "world"
	}
	return "hi " + name, nil
}

func (g *greetAPI) McpServiceDesc() *ServiceDesc {
	return &ServiceDesc{
		ServiceName: "api.v1.Greet",
		Name:        "Greet",
		HandlerType: (*greetServer)(nil),
		Tools: []ToolDesc{
			{
				Name:        "greet",
				Description: "say hi",
				InputSchema: json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`),
				Handler: func(ctx context.Context, req *mcpsdk.CallToolRequest, srv any, interceptor server.UnaryServerInterceptor) (*mcpsdk.CallToolResult, error) {
					var in struct {
						Name string `json:"name"`
					}
					if len(req.Params.Arguments) > 0 {
						if err := json.Unmarshal(req.Params.Arguments, &in); err != nil {
							return nil, err
						}
					}
					info := &server.UnaryServerInfo{
						Server:     srv,
						FullMethod: "/api.v1.Greet/Greet",
						Protocol:   Protocol,
					}
					handler := func(ctx context.Context, req any) (any, error) {
						return srv.(greetServer).Greet(ctx, in.Name)
					}
					if interceptor == nil {
						interceptor = passthrough
					}
					resp, err := interceptor(ctx, in, info, handler)
					if err != nil {
						return nil, err
					}
					return &mcpsdk.CallToolResult{
						Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: resp.(string)}},
					}, nil
				},
			},
		},
	}
}

func passthrough(ctx context.Context, req any, _ *server.UnaryServerInfo, handler server.UnaryHandler) (any, error) {
	return handler(ctx, req)
}

type mcpHandler struct {
	desc *ServiceDesc
}

func (h mcpHandler) McpServiceDesc() *ServiceDesc { return h.desc }

func TestDefaultConfig(t *testing.T) {
	conf := defaultConfig()
	require.Equal(t, DefaultPath, conf.Path)
	require.Equal(t, server.GetConfigWithProtocol(Protocol), conf.Config.Config.Config)
	require.Zero(t, conf.Options.ReadTimeout.Duration)
	require.Zero(t, conf.Options.ReadHeaderTimeout.Duration)
	require.Zero(t, conf.Options.IdleTimeout.Duration)
}

func TestMcpServerContract(t *testing.T) {
	conf := defaultConfig()
	conf.Enabled = true
	created, err := MustNew(conf, &server.ServerOptions{})
	require.NoError(t, err)
	s := created.(*McpServer)
	require.Equal(t, Protocol, s.Protocol())
	require.True(t, s.Enabled())
	require.Equal(t, conf.Addresses, s.ListenAddresses())
	require.Error(t, s.AddHandler(struct{}{}))
	require.Error(t, s.Start(make(chan error, 1)))
	require.NoError(t, s.AddHandler(&greetAPI{}))
	s.Stop()
}

func TestAddHandlerRejectsMismatchedType(t *testing.T) {
	type other interface{ Other() }
	created, err := MustNew(defaultConfig(), &server.ServerOptions{})
	require.NoError(t, err)
	s := created.(*McpServer)
	err = s.AddHandler(mcpHandler{desc: &ServiceDesc{
		HandlerType: (*other)(nil),
		Tools: []ToolDesc{{
			Name: "x",
			Handler: func(context.Context, *mcpsdk.CallToolRequest, any, server.UnaryServerInterceptor) (*mcpsdk.CallToolResult, error) {
				return nil, nil
			},
		}},
	}})
	require.Error(t, err)
}

func TestMcpServerOverHTTP(t *testing.T) {
	conf := defaultConfig()
	conf.Enabled = true
	conf.Addresses.Listen = freeListenAddr(t)
	created, err := MustNew(conf, &server.ServerOptions{})
	require.NoError(t, err)
	s := created.(*McpServer)
	require.NoError(t, s.AddHandler(&greetAPI{}))

	startErr := make(chan error, 1)
	require.NoError(t, s.Start(startErr))
	t.Cleanup(s.Stop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	session := connectMCP(t, ctx, "http://"+conf.Addresses.Listen+conf.Path)
	t.Cleanup(func() { _ = session.Close() })

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 1)
	require.Equal(t, "greet", tools.Tools[0].Name)

	result, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "greet",
		Arguments: map[string]any{"name": "asjard"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	text, ok := result.Content[0].(*mcpsdk.TextContent)
	require.True(t, ok)
	require.Equal(t, "hi asjard", text.Text)

	select {
	case err := <-startErr:
		require.NoError(t, err)
	default:
	}
}

func freeListenAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	return addr
}

func connectMCP(t *testing.T, ctx context.Context, endpoint string) *mcpsdk.ClientSession {
	t.Helper()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "v1"}, nil)
	deadline := time.Now().Add(2 * time.Second)
	var session *mcpsdk.ClientSession
	var err error
	for {
		session, err = client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: endpoint}, nil)
		if err == nil {
			return session
		}
		if time.Now().After(deadline) {
			require.NoError(t, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
