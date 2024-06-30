// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-rest v1.3.0
// - protoc             v5.27.0
// source: hello/hello.proto

package hello

import (
	context "context"
	server "github.com/asjard/asjard/core/server"
	rest "github.com/asjard/asjard/pkg/server/rest"
)

func _Hello_Say_RestHandler(ctx *rest.Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
	in := new(SayReq)
	if interceptor == nil {
		return srv.(HelloServer).Say(ctx, in)
	}
	info := &server.UnaryServerInfo{
		Server:     srv,
		FullMethod: "api.v1.hello.Hello.Say",
		Protocol:   rest.Protocol,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(HelloServer).Say(ctx, in)
	}
	return interceptor(ctx, in, info, handler)
}
func _Hello_Call_RestHandler(ctx *rest.Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
	in := new(SayReq)
	if interceptor == nil {
		return srv.(HelloServer).Call(ctx, in)
	}
	info := &server.UnaryServerInfo{
		Server:     srv,
		FullMethod: "api.v1.hello.Hello.Call",
		Protocol:   rest.Protocol,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(HelloServer).Call(ctx, in)
	}
	return interceptor(ctx, in, info, handler)
}
func _Hello_CipherExample_RestHandler(ctx *rest.Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
	in := new(CipherExampleReq)
	if interceptor == nil {
		return srv.(HelloServer).CipherExample(ctx, in)
	}
	info := &server.UnaryServerInfo{
		Server:     srv,
		FullMethod: "api.v1.hello.Hello.CipherExample",
		Protocol:   rest.Protocol,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(HelloServer).CipherExample(ctx, in)
	}
	return interceptor(ctx, in, info, handler)
}
func _Hello_MysqlExample_RestHandler(ctx *rest.Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
	in := new(MysqlExampleReq)
	if interceptor == nil {
		return srv.(HelloServer).MysqlExample(ctx, in)
	}
	info := &server.UnaryServerInfo{
		Server:     srv,
		FullMethod: "api.v1.hello.Hello.MysqlExample",
		Protocol:   rest.Protocol,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(HelloServer).MysqlExample(ctx, in)
	}
	return interceptor(ctx, in, info, handler)
}

// HelloRestServiceDesc is the rest.ServiceDesc for Hello service.
// It's only intended for direct use with rest.AddHandler,
// and not to be introspected or modified (even as a copy)
var HelloRestServiceDesc = rest.ServiceDesc{
	ServiceName: "api.v1.hello.Hello",
	HandlerType: (*HelloServer)(nil),
	Methods: []rest.MethodDesc{
		{
			MethodName: "Say",
			Desc:       "say something.",
			Method:     "POST",
			Path:       "/api/v1/region/{region_id}/project/{project_id}/user/{user_id}",
			Handler:    _Hello_Say_RestHandler,
		},
		{
			MethodName: "Call",
			Desc:       ".",
			Method:     "POST",
			Path:       "/api/v1/call/region/{region_id}/project/{project_id}/user/{user_id}",
			Handler:    _Hello_Call_RestHandler,
		},
		{
			MethodName: "CipherExample",
			Desc:       "加解密示例.",
			Method:     "GET",
			Path:       "/api/v1/examples/cipher",
			Handler:    _Hello_CipherExample_RestHandler,
		},
		{
			MethodName: "MysqlExample",
			Desc:       "mysql数据库示例.",
			Method:     "POST",
			Path:       "/api/v1/examples/mysql",
			Handler:    _Hello_MysqlExample_RestHandler,
		},
	},
}
