package mcp

import (
	"encoding/json"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// CallToolRequest is an alias of the MCP SDK request type for generated handlers.
type CallToolRequest = mcpsdk.CallToolRequest

// CallToolResult is an alias of the MCP SDK result type for generated handlers.
type CallToolResult = mcpsdk.CallToolResult

// UnmarshalToolArguments decodes MCP tool arguments into a protobuf message.
func UnmarshalToolArguments(req *mcpsdk.CallToolRequest, in proto.Message) error {
	if req == nil || req.Params == nil || len(req.Params.Arguments) == 0 {
		return nil
	}
	return protojson.Unmarshal(req.Params.Arguments, in)
}

// NewProtoToolResult marshals a protobuf message into an MCP tool result.
// Business errors are returned via CallToolResult.SetError instead of a protocol error.
func NewProtoToolResult(out proto.Message, err error) (*mcpsdk.CallToolResult, error) {
	result := &mcpsdk.CallToolResult{}
	if err != nil {
		result.SetError(err)
		return result, nil
	}
	if out == nil {
		return result, nil
	}
	b, marshalErr := protojson.Marshal(out)
	if marshalErr != nil {
		result.SetError(marshalErr)
		return result, nil
	}
	result.Content = []mcpsdk.Content{&mcpsdk.TextContent{Text: string(b)}}
	result.StructuredContent = json.RawMessage(b)
	return result, nil
}

// NewProtoToolResultFromAny converts an interceptor response into an MCP tool result.
func NewProtoToolResultFromAny(resp any, err error) (*mcpsdk.CallToolResult, error) {
	if err != nil {
		return NewProtoToolResult(nil, err)
	}
	if resp == nil {
		return NewProtoToolResult(nil, nil)
	}
	out, ok := resp.(proto.Message)
	if !ok {
		return NewProtoToolResult(nil, fmt.Errorf("unexpected mcp response type %T", resp))
	}
	return NewProtoToolResult(out, nil)
}
