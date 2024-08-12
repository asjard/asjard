// Code generated by protoc-gen-go-rest. DO NOT EDIT.
// versions:
// - protoc-gen-go-rest v1.3.0
// - protoc             v5.27.0
// source: openapi.proto

package rest

import (
	context "context"
	server "github.com/asjard/asjard/core/server"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// 获取openapi yaml文件
func _OpenAPI_Yaml_RestHandler(ctx *Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
	in := new(emptypb.Empty)
	if interceptor == nil {
		return srv.(OpenAPIServer).Yaml(ctx, in)
	}
	info := &server.UnaryServerInfo{
		Server:     srv,
		FullMethod: OpenAPI_Yaml_FullMethodName,
		Protocol:   Protocol,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(OpenAPIServer).Yaml(ctx, in)
	}
	return interceptor(ctx, in, info, handler)
}

// 文档页面
func _OpenAPI_Page_RestHandler(ctx *Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
	in := new(emptypb.Empty)
	if interceptor == nil {
		return srv.(OpenAPIServer).Page(ctx, in)
	}
	info := &server.UnaryServerInfo{
		Server:     srv,
		FullMethod: OpenAPI_Page_FullMethodName,
		Protocol:   Protocol,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(OpenAPIServer).Page(ctx, in)
	}
	return interceptor(ctx, in, info, handler)
}

// OpenAPIRestServiceDesc is the ServiceDesc for OpenAPI service.
// It's only intended for direct use with AddHandler,
// and not to be introspected or modified (even as a copy)
//
// OpenAPI相关文档
var OpenAPIRestServiceDesc = ServiceDesc{
	ServiceName: "asjard.api.OpenAPI",
	HandlerType: (*OpenAPIServer)(nil),
	OpenAPI:     file_openapi_proto_openapi,
	Methods: []MethodDesc{
		{
			MethodName: "Yaml",
			Desc:       "获取openapi yaml文件.",
			Method:     "GET",
			Path:       OpenAPI_Yaml_RestPath,
			Handler:    _OpenAPI_Yaml_RestHandler,
		},
		{
			MethodName: "Page",
			Desc:       "文档页面.",
			Method:     "GET",
			Path:       OpenAPI_Page_RestPath,
			Handler:    _OpenAPI_Page_RestHandler,
		},
	},
}

const (
	OpenAPI_Yaml_RestPath = "/openapi.yml"
	OpenAPI_Page_RestPath = "/page/openapi"
)

var file_openapi_proto_openapi = []byte{
	0x0a, 0x05, 0x33, 0x2e, 0x30, 0x2e, 0x33, 0x12, 0x07, 0x32, 0x05, 0x30, 0x2e, 0x30, 0x2e, 0x31,
	0x22, 0x85, 0x03, 0x0a, 0xc5, 0x01, 0x0a, 0x0c, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69,
	0x2e, 0x79, 0x6d, 0x6c, 0x12, 0xb4, 0x01, 0x22, 0xb1, 0x01, 0x0a, 0x07, 0x4f, 0x70, 0x65, 0x6e,
	0x41, 0x50, 0x49, 0x1a, 0x18, 0xe8, 0x8e, 0xb7, 0xe5, 0x8f, 0x96, 0x6f, 0x70, 0x65, 0x6e, 0x61,
	0x70, 0x69, 0x20, 0x79, 0x61, 0x6d, 0x6c, 0xe6, 0x96, 0x87, 0xe4, 0xbb, 0xb6, 0x2a, 0x19, 0x61,
	0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4f, 0x70, 0x65, 0x6e, 0x41, 0x50,
	0x49, 0x2e, 0x59, 0x61, 0x6d, 0x6c, 0x5f, 0x30, 0x42, 0x71, 0x12, 0x0f, 0x0a, 0x03, 0x32, 0x30,
	0x30, 0x12, 0x08, 0x0a, 0x06, 0x0a, 0x02, 0x4f, 0x4b, 0x1a, 0x00, 0x12, 0x5e, 0x0a, 0x07, 0x64,
	0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x53, 0x0a, 0x51, 0x0a, 0x16, 0x44, 0x65, 0x66, 0x61,
	0x75, 0x6c, 0x74, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x1a, 0x37, 0x0a, 0x35, 0x0a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x1f, 0x12, 0x1d, 0x0a, 0x1b,
	0x23, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x73, 0x2f, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x0a, 0xba, 0x01, 0x0a, 0x0d,
	0x2f, 0x70, 0x61, 0x67, 0x65, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x12, 0xa8, 0x01,
	0x22, 0xa5, 0x01, 0x0a, 0x07, 0x4f, 0x70, 0x65, 0x6e, 0x41, 0x50, 0x49, 0x1a, 0x0c, 0xe6, 0x96,
	0x87, 0xe6, 0xa1, 0xa3, 0xe9, 0xa1, 0xb5, 0xe9, 0x9d, 0xa2, 0x2a, 0x19, 0x61, 0x73, 0x6a, 0x61,
	0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4f, 0x70, 0x65, 0x6e, 0x41, 0x50, 0x49, 0x2e, 0x50,
	0x61, 0x67, 0x65, 0x5f, 0x30, 0x42, 0x71, 0x12, 0x0f, 0x0a, 0x03, 0x32, 0x30, 0x30, 0x12, 0x08,
	0x0a, 0x06, 0x0a, 0x02, 0x4f, 0x4b, 0x1a, 0x00, 0x12, 0x5e, 0x0a, 0x07, 0x64, 0x65, 0x66, 0x61,
	0x75, 0x6c, 0x74, 0x12, 0x53, 0x0a, 0x51, 0x0a, 0x16, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74,
	0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x1a,
	0x37, 0x0a, 0x35, 0x0a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x1f, 0x12, 0x1d, 0x0a, 0x1b, 0x23, 0x2f, 0x63,
	0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x73, 0x2f, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a, 0x97, 0x09, 0x0a, 0x94, 0x09, 0x0a, 0xd7,
	0x01, 0x0a, 0x11, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x41, 0x6e, 0x79, 0x12, 0xc1, 0x01, 0x0a, 0xbe, 0x01, 0xca, 0x01, 0x06, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0xfa, 0x01, 0x3c, 0x0a, 0x3a, 0x0a, 0x05, 0x40, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x31, 0x0a, 0x2f, 0xca, 0x01, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x92, 0x02, 0x23, 0x54,
	0x68, 0x65, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x6f, 0x66, 0x20, 0x74, 0x68, 0x65, 0x20, 0x73,
	0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x2e, 0x82, 0x02, 0x02, 0x10, 0x01, 0x92, 0x02, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x73, 0x20, 0x61, 0x6e, 0x20, 0x61, 0x72, 0x62, 0x69, 0x74, 0x72, 0x61, 0x72, 0x79, 0x20,
	0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x20, 0x61, 0x6c, 0x6f, 0x6e, 0x67, 0x20, 0x77, 0x69, 0x74, 0x68, 0x20, 0x61, 0x20,
	0x40, 0x74, 0x79, 0x70, 0x65, 0x20, 0x74, 0x68, 0x61, 0x74, 0x20, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x62, 0x65, 0x73, 0x20, 0x74, 0x68, 0x65, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x6f, 0x66,
	0x20, 0x74, 0x68, 0x65, 0x20, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x20,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x0a, 0xb7, 0x07, 0x0a, 0x06, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0xac, 0x07, 0x0a, 0xa9, 0x07, 0xca, 0x01, 0x06, 0x6f, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0xfa, 0x01, 0xef, 0x03, 0x0a, 0x74, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x6c,
	0x0a, 0x6a, 0xca, 0x01, 0x07, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x65, 0x72, 0x92, 0x02, 0x55, 0x54,
	0x68, 0x65, 0x20, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2c, 0x20,
	0x77, 0x68, 0x69, 0x63, 0x68, 0x20, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x20, 0x62, 0x65, 0x20,
	0x61, 0x6e, 0x20, 0x65, 0x6e, 0x75, 0x6d, 0x20, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x20, 0x6f, 0x66,
	0x20, 0x5b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x64,
	0x65, 0x5d, 0x5b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f,
	0x64, 0x65, 0x5d, 0x2e, 0x9a, 0x02, 0x05, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x0a, 0xf5, 0x01, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0xe9, 0x01, 0x0a, 0xe6, 0x01, 0xca, 0x01,
	0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x92, 0x02, 0xd9, 0x01, 0x41, 0x20, 0x64, 0x65, 0x76,
	0x65, 0x6c, 0x6f, 0x70, 0x65, 0x72, 0x2d, 0x66, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2c, 0x20, 0x77, 0x68, 0x69,
	0x63, 0x68, 0x20, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x20, 0x62, 0x65, 0x20, 0x69, 0x6e, 0x20,
	0x45, 0x6e, 0x67, 0x6c, 0x69, 0x73, 0x68, 0x2e, 0x20, 0x41, 0x6e, 0x79, 0x20, 0x75, 0x73, 0x65,
	0x72, 0x2d, 0x66, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x20, 0x62, 0x65,
	0x20, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x73,
	0x65, 0x6e, 0x74, 0x20, 0x69, 0x6e, 0x20, 0x74, 0x68, 0x65, 0x20, 0x5b, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x64, 0x65,
	0x74, 0x61, 0x69, 0x6c, 0x73, 0x5d, 0x5b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70,
	0x63, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73,
	0x5d, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x2c, 0x20, 0x6f, 0x72, 0x20, 0x6c, 0x6f, 0x63, 0x61,
	0x6c, 0x69, 0x7a, 0x65, 0x64, 0x20, 0x62, 0x79, 0x20, 0x74, 0x68, 0x65, 0x20, 0x63, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x2e, 0x0a, 0x7f, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x77, 0x0a, 0x75,
	0xca, 0x01, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x92, 0x02, 0x69, 0x41, 0x20, 0x6c, 0x69,
	0x73, 0x74, 0x20, 0x6f, 0x66, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x20, 0x74,
	0x68, 0x61, 0x74, 0x20, 0x63, 0x61, 0x72, 0x72, 0x79, 0x20, 0x74, 0x68, 0x65, 0x20, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x20, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x2e, 0x20, 0x20, 0x54, 0x68,
	0x65, 0x72, 0x65, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x20,
	0x73, 0x65, 0x74, 0x20, 0x6f, 0x66, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x20, 0x66, 0x6f, 0x72, 0x20, 0x41, 0x50, 0x49, 0x73, 0x20, 0x74, 0x6f,
	0x20, 0x75, 0x73, 0x65, 0x2e, 0x92, 0x02, 0xa9, 0x03, 0x54, 0x68, 0x65, 0x20, 0x60, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x60, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x64, 0x65, 0x66, 0x69, 0x6e,
	0x65, 0x73, 0x20, 0x61, 0x20, 0x6c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c, 0x20, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x20, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x20, 0x74, 0x68, 0x61, 0x74, 0x20, 0x69, 0x73,
	0x20, 0x73, 0x75, 0x69, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x20, 0x66, 0x6f, 0x72, 0x20, 0x64, 0x69,
	0x66, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x74, 0x20, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x61, 0x6d, 0x6d,
	0x69, 0x6e, 0x67, 0x20, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x2c, 0x20, 0x69, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x69, 0x6e, 0x67, 0x20, 0x52, 0x45, 0x53, 0x54,
	0x20, 0x41, 0x50, 0x49, 0x73, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x52, 0x50, 0x43, 0x20, 0x41, 0x50,
	0x49, 0x73, 0x2e, 0x20, 0x49, 0x74, 0x20, 0x69, 0x73, 0x20, 0x75, 0x73, 0x65, 0x64, 0x20, 0x62,
	0x79, 0x20, 0x5b, 0x67, 0x52, 0x50, 0x43, 0x5d, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f,
	0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x72, 0x70, 0x63,
	0x29, 0x2e, 0x20, 0x45, 0x61, 0x63, 0x68, 0x20, 0x60, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x60,
	0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e,
	0x73, 0x20, 0x74, 0x68, 0x72, 0x65, 0x65, 0x20, 0x70, 0x69, 0x65, 0x63, 0x65, 0x73, 0x20, 0x6f,
	0x66, 0x20, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x63, 0x6f,
	0x64, 0x65, 0x2c, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x2c, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x64, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x73, 0x2e, 0x20, 0x59, 0x6f, 0x75, 0x20, 0x63, 0x61, 0x6e, 0x20, 0x66, 0x69,
	0x6e, 0x64, 0x20, 0x6f, 0x75, 0x74, 0x20, 0x6d, 0x6f, 0x72, 0x65, 0x20, 0x61, 0x62, 0x6f, 0x75,
	0x74, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x68, 0x6f, 0x77, 0x20, 0x74, 0x6f, 0x20, 0x77, 0x6f,
	0x72, 0x6b, 0x20, 0x77, 0x69, 0x74, 0x68, 0x20, 0x69, 0x74, 0x20, 0x69, 0x6e, 0x20, 0x74, 0x68,
	0x65, 0x20, 0x5b, 0x41, 0x50, 0x49, 0x20, 0x44, 0x65, 0x73, 0x69, 0x67, 0x6e, 0x20, 0x47, 0x75,
	0x69, 0x64, 0x65, 0x5d, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70,
	0x69, 0x73, 0x2f, 0x64, 0x65, 0x73, 0x69, 0x67, 0x6e, 0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73,
	0x29, 0x2e,
}
