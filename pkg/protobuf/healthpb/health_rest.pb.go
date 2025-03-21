// Code generated by protoc-gen-go-rest. DO NOT EDIT.
// versions:
// - protoc-gen-go-rest v1.3.0
// - protoc             v5.27.0
// source: health.proto

package healthpb

import (
	context "context"
	server "github.com/asjard/asjard/core/server"
	rest "github.com/asjard/asjard/pkg/server/rest"
)

// HealthCheck.
func _Health_Check_RestHandler(ctx *rest.Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
	in := new(HealthCheckRequest)
	if interceptor == nil {
		return srv.(HealthServer).Check(ctx, in)
	}
	info := &server.UnaryServerInfo{
		Server:     srv,
		FullMethod: Health_Check_FullMethodName,
		Protocol:   rest.Protocol,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(HealthServer).Check(ctx, in)
	}
	return interceptor(ctx, in, info, handler)
}

// HealthRestServiceDesc is the rest.ServiceDesc for Health service.
// It's only intended for direct use with rest.AddHandler,
// and not to be introspected or modified (even as a copy)
var HealthRestServiceDesc = rest.ServiceDesc{
	ServiceName: "asjard.api.health.Health",
	HandlerType: (*HealthServer)(nil),
	OpenAPI:     file_health_proto_openapi,
	Methods: []rest.MethodDesc{
		{
			MethodName: "Check",
			Desc:       "HealthCheck..",
			Method:     "GET",
			Path:       Health_Check_RestPath,
			Handler:    _Health_Check_RestHandler,
		},
	},
}

const (
	Health_Check_RestPath = "/health"
)

var file_health_proto_openapi = []byte{
	0x0a, 0x05, 0x33, 0x2e, 0x30, 0x2e, 0x33, 0x12, 0x07, 0x32, 0x05, 0x30, 0x2e, 0x30, 0x2e, 0x31,
	0x22, 0xd3, 0x02, 0x0a, 0xd0, 0x02, 0x0a, 0x07, 0x2f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12,
	0xc4, 0x02, 0x22, 0xc1, 0x02, 0x0a, 0x0b, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x48, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x12, 0x0c, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x20, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x1a, 0x0c, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x2e, 0x2a, 0x20,
	0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x30,
	0x32, 0x1f, 0x0a, 0x1d, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x05, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x52, 0x0b, 0x0a, 0x09, 0xca, 0x01, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x42, 0xd2, 0x01, 0x12, 0x65, 0x0a, 0x03, 0x32, 0x30, 0x30, 0x12, 0x5e, 0x0a, 0x5c, 0x0a,
	0x02, 0x4f, 0x4b, 0x1a, 0x56, 0x0a, 0x54, 0x0a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x40, 0x0a, 0x3e, 0x12, 0x3c, 0x0a,
	0x3a, 0x23, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x73, 0x2f, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x69, 0x0a, 0x07, 0x64,
	0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x5e, 0x0a, 0x5c, 0x0a, 0x16, 0x44, 0x65, 0x66, 0x61,
	0x75, 0x6c, 0x74, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x1a, 0x42, 0x0a, 0x40, 0x0a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x2c, 0x0a, 0x2a, 0x12, 0x28, 0x0a, 0x26,
	0x23, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x73, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a, 0xbc, 0x0a, 0x0a, 0xb9, 0x0a, 0x0a, 0x95, 0x01, 0x0a,
	0x25, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x68, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6c, 0x0a, 0x6a, 0xca, 0x01, 0x06, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0xfa, 0x01, 0x5e, 0x0a, 0x44, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x3a, 0x0a, 0x38, 0xc2, 0x01, 0x09, 0x12, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e,
	0xc2, 0x01, 0x09, 0x12, 0x07, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e, 0x47, 0xc2, 0x01, 0x0d, 0x12,
	0x0b, 0x4e, 0x4f, 0x54, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e, 0x47, 0xca, 0x01, 0x06, 0x73,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x9a, 0x02, 0x04, 0x65, 0x6e, 0x75, 0x6d, 0x0a, 0x16, 0x0a, 0x07,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x0b, 0x0a, 0x09, 0xca, 0x01, 0x06, 0x73, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x0a, 0xd9, 0x01, 0x0a, 0x13, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x12, 0xc1, 0x01, 0x0a,
	0xbe, 0x01, 0xca, 0x01, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0xfa, 0x01, 0x3c, 0x0a, 0x3a,
	0x0a, 0x05, 0x40, 0x74, 0x79, 0x70, 0x65, 0x12, 0x31, 0x0a, 0x2f, 0xca, 0x01, 0x06, 0x73, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x92, 0x02, 0x23, 0x54, 0x68, 0x65, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20,
	0x6f, 0x66, 0x20, 0x74, 0x68, 0x65, 0x20, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65,
	0x64, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x82, 0x02, 0x02, 0x10, 0x01, 0x92,
	0x02, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x73, 0x20, 0x61, 0x6e, 0x20, 0x61, 0x72,
	0x62, 0x69, 0x74, 0x72, 0x61, 0x72, 0x79, 0x20, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a,
	0x65, 0x64, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20, 0x61, 0x6c, 0x6f, 0x6e, 0x67,
	0x20, 0x77, 0x69, 0x74, 0x68, 0x20, 0x61, 0x20, 0x40, 0x74, 0x79, 0x70, 0x65, 0x20, 0x74, 0x68,
	0x61, 0x74, 0x20, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x73, 0x20, 0x74, 0x68, 0x65,
	0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x6f, 0x66, 0x20, 0x74, 0x68, 0x65, 0x20, 0x73, 0x65, 0x72,
	0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e,
	0x0a, 0xc2, 0x07, 0x0a, 0x11, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0xac, 0x07, 0x0a, 0xa9, 0x07, 0xca, 0x01, 0x06, 0x6f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0xfa, 0x01, 0xef, 0x03, 0x0a, 0x74, 0x0a, 0x04, 0x63, 0x6f, 0x64,
	0x65, 0x12, 0x6c, 0x0a, 0x6a, 0xca, 0x01, 0x07, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x65, 0x72, 0x92,
	0x02, 0x55, 0x54, 0x68, 0x65, 0x20, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x20, 0x63, 0x6f, 0x64,
	0x65, 0x2c, 0x20, 0x77, 0x68, 0x69, 0x63, 0x68, 0x20, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x20,
	0x62, 0x65, 0x20, 0x61, 0x6e, 0x20, 0x65, 0x6e, 0x75, 0x6d, 0x20, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x20, 0x6f, 0x66, 0x20, 0x5b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x43, 0x6f, 0x64, 0x65, 0x5d, 0x5b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63,
	0x2e, 0x43, 0x6f, 0x64, 0x65, 0x5d, 0x2e, 0x9a, 0x02, 0x05, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x0a,
	0xf5, 0x01, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0xe9, 0x01, 0x0a, 0xe6,
	0x01, 0xca, 0x01, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x92, 0x02, 0xd9, 0x01, 0x41, 0x20,
	0x64, 0x65, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x72, 0x2d, 0x66, 0x61, 0x63, 0x69, 0x6e, 0x67,
	0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2c, 0x20,
	0x77, 0x68, 0x69, 0x63, 0x68, 0x20, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x20, 0x62, 0x65, 0x20,
	0x69, 0x6e, 0x20, 0x45, 0x6e, 0x67, 0x6c, 0x69, 0x73, 0x68, 0x2e, 0x20, 0x41, 0x6e, 0x79, 0x20,
	0x75, 0x73, 0x65, 0x72, 0x2d, 0x66, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64,
	0x20, 0x62, 0x65, 0x20, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x20, 0x61, 0x6e,
	0x64, 0x20, 0x73, 0x65, 0x6e, 0x74, 0x20, 0x69, 0x6e, 0x20, 0x74, 0x68, 0x65, 0x20, 0x5b, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x2e, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x5d, 0x5b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x64, 0x65, 0x74, 0x61,
	0x69, 0x6c, 0x73, 0x5d, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x2c, 0x20, 0x6f, 0x72, 0x20, 0x6c,
	0x6f, 0x63, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x20, 0x62, 0x79, 0x20, 0x74, 0x68, 0x65, 0x20,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x0a, 0x7f, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x77, 0x0a, 0x75, 0xca, 0x01, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x92, 0x02, 0x69, 0x41,
	0x20, 0x6c, 0x69, 0x73, 0x74, 0x20, 0x6f, 0x66, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x73, 0x20, 0x74, 0x68, 0x61, 0x74, 0x20, 0x63, 0x61, 0x72, 0x72, 0x79, 0x20, 0x74, 0x68, 0x65,
	0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x2e, 0x20,
	0x20, 0x54, 0x68, 0x65, 0x72, 0x65, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x20, 0x73, 0x65, 0x74, 0x20, 0x6f, 0x66, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x20, 0x74, 0x79, 0x70, 0x65, 0x73, 0x20, 0x66, 0x6f, 0x72, 0x20, 0x41, 0x50, 0x49, 0x73,
	0x20, 0x74, 0x6f, 0x20, 0x75, 0x73, 0x65, 0x2e, 0x92, 0x02, 0xa9, 0x03, 0x54, 0x68, 0x65, 0x20,
	0x60, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x60, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x64, 0x65,
	0x66, 0x69, 0x6e, 0x65, 0x73, 0x20, 0x61, 0x20, 0x6c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c, 0x20,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x20, 0x74, 0x68, 0x61, 0x74,
	0x20, 0x69, 0x73, 0x20, 0x73, 0x75, 0x69, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x20, 0x66, 0x6f, 0x72,
	0x20, 0x64, 0x69, 0x66, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x74, 0x20, 0x70, 0x72, 0x6f, 0x67, 0x72,
	0x61, 0x6d, 0x6d, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65,
	0x6e, 0x74, 0x73, 0x2c, 0x20, 0x69, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x69, 0x6e, 0x67, 0x20, 0x52,
	0x45, 0x53, 0x54, 0x20, 0x41, 0x50, 0x49, 0x73, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x52, 0x50, 0x43,
	0x20, 0x41, 0x50, 0x49, 0x73, 0x2e, 0x20, 0x49, 0x74, 0x20, 0x69, 0x73, 0x20, 0x75, 0x73, 0x65,
	0x64, 0x20, 0x62, 0x79, 0x20, 0x5b, 0x67, 0x52, 0x50, 0x43, 0x5d, 0x28, 0x68, 0x74, 0x74, 0x70,
	0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67,
	0x72, 0x70, 0x63, 0x29, 0x2e, 0x20, 0x45, 0x61, 0x63, 0x68, 0x20, 0x60, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x60, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x73, 0x20, 0x74, 0x68, 0x72, 0x65, 0x65, 0x20, 0x70, 0x69, 0x65, 0x63, 0x65,
	0x73, 0x20, 0x6f, 0x66, 0x20, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x20, 0x63, 0x6f, 0x64, 0x65, 0x2c, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x2c, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20,
	0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x2e, 0x20, 0x59, 0x6f, 0x75, 0x20, 0x63, 0x61, 0x6e,
	0x20, 0x66, 0x69, 0x6e, 0x64, 0x20, 0x6f, 0x75, 0x74, 0x20, 0x6d, 0x6f, 0x72, 0x65, 0x20, 0x61,
	0x62, 0x6f, 0x75, 0x74, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x20,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x68, 0x6f, 0x77, 0x20, 0x74, 0x6f,
	0x20, 0x77, 0x6f, 0x72, 0x6b, 0x20, 0x77, 0x69, 0x74, 0x68, 0x20, 0x69, 0x74, 0x20, 0x69, 0x6e,
	0x20, 0x74, 0x68, 0x65, 0x20, 0x5b, 0x41, 0x50, 0x49, 0x20, 0x44, 0x65, 0x73, 0x69, 0x67, 0x6e,
	0x20, 0x47, 0x75, 0x69, 0x64, 0x65, 0x5d, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x64, 0x65, 0x73, 0x69, 0x67, 0x6e, 0x2f, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x73, 0x29, 0x2e,
}
