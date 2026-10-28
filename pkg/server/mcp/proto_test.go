package mcp

import (
	"encoding/json"
	"errors"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestUnmarshalToolArguments(t *testing.T) {
	in := &wrapperspb.StringValue{}
	require.NoError(t, UnmarshalToolArguments(nil, in))
	require.NoError(t, UnmarshalToolArguments(&mcpsdk.CallToolRequest{}, in))

	raw, err := protojson.Marshal(&wrapperspb.StringValue{Value: "hello"})
	require.NoError(t, err)
	req := &mcpsdk.CallToolRequest{Params: &mcpsdk.CallToolParamsRaw{Arguments: raw}}
	require.NoError(t, UnmarshalToolArguments(req, in))
	require.Equal(t, "hello", in.GetValue())

	req.Params.Arguments = []byte(`{`)
	require.Error(t, UnmarshalToolArguments(req, in))
}

func TestNewProtoToolResult(t *testing.T) {
	empty, err := NewProtoToolResult(nil, nil)
	require.NoError(t, err)
	require.Empty(t, empty.Content)
	require.False(t, empty.IsError)

	result, err := NewProtoToolResult(&wrapperspb.StringValue{Value: "hi"}, nil)
	require.NoError(t, err)
	require.False(t, result.IsError)
	text, ok := result.Content[0].(*mcpsdk.TextContent)
	require.True(t, ok)
	require.Contains(t, text.Text, "hi")
	raw, ok := result.StructuredContent.(json.RawMessage)
	require.True(t, ok)
	require.JSONEq(t, text.Text, string(raw))

	failed, err := NewProtoToolResult(nil, errors.New("boom"))
	require.NoError(t, err)
	require.True(t, failed.IsError)
	require.Equal(t, "boom", failed.GetError().Error())
	failText, ok := failed.Content[0].(*mcpsdk.TextContent)
	require.True(t, ok)
	require.Equal(t, "boom", failText.Text)

	fromAny, err := NewProtoToolResultFromAny(&wrapperspb.StringValue{Value: "ok"}, nil)
	require.NoError(t, err)
	require.False(t, fromAny.IsError)
	require.NotEmpty(t, fromAny.Content)

	bad, err := NewProtoToolResultFromAny("bad", nil)
	require.NoError(t, err)
	require.True(t, bad.IsError)

	wrapped, err := NewProtoToolResultFromAny(&wrapperspb.StringValue{Value: "ok"}, errors.New("interceptor"))
	require.NoError(t, err)
	require.True(t, wrapped.IsError)
	require.Equal(t, "interceptor", wrapped.GetError().Error())
}
