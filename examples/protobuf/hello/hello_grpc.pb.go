// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.27.0
// source: hello/hello.proto

package hello

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Hello_Say_FullMethodName           = "/api.v1.hello.Hello/Say"
	Hello_Call_FullMethodName          = "/api.v1.hello.Hello/Call"
	Hello_Log_FullMethodName           = "/api.v1.hello.Hello/Log"
	Hello_CipherExample_FullMethodName = "/api.v1.hello.Hello/CipherExample"
	Hello_MysqlExample_FullMethodName  = "/api.v1.hello.Hello/MysqlExample"
)

// HelloClient is the client API for Hello service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HelloClient interface {
	// say something
	Say(ctx context.Context, in *SayReq, opts ...grpc.CallOption) (*SayReq, error)
	Call(ctx context.Context, in *SayReq, opts ...grpc.CallOption) (*SayReq, error)
	// 获取日志
	Log(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// 加解密示例
	CipherExample(ctx context.Context, in *CipherExampleReq, opts ...grpc.CallOption) (*CipherExampleResp, error)
	// mysql数据库示例
	MysqlExample(ctx context.Context, in *MysqlExampleReq, opts ...grpc.CallOption) (*MysqlExampleResp, error)
}

type helloClient struct {
	cc grpc.ClientConnInterface
}

func NewHelloClient(cc grpc.ClientConnInterface) HelloClient {
	return &helloClient{cc}
}

func (c *helloClient) Say(ctx context.Context, in *SayReq, opts ...grpc.CallOption) (*SayReq, error) {
	out := new(SayReq)
	err := c.cc.Invoke(ctx, Hello_Say_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helloClient) Call(ctx context.Context, in *SayReq, opts ...grpc.CallOption) (*SayReq, error) {
	out := new(SayReq)
	err := c.cc.Invoke(ctx, Hello_Call_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helloClient) Log(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Hello_Log_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helloClient) CipherExample(ctx context.Context, in *CipherExampleReq, opts ...grpc.CallOption) (*CipherExampleResp, error) {
	out := new(CipherExampleResp)
	err := c.cc.Invoke(ctx, Hello_CipherExample_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helloClient) MysqlExample(ctx context.Context, in *MysqlExampleReq, opts ...grpc.CallOption) (*MysqlExampleResp, error) {
	out := new(MysqlExampleResp)
	err := c.cc.Invoke(ctx, Hello_MysqlExample_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HelloServer is the server API for Hello service.
// All implementations must embed UnimplementedHelloServer
// for forward compatibility
type HelloServer interface {
	// say something
	Say(context.Context, *SayReq) (*SayReq, error)
	Call(context.Context, *SayReq) (*SayReq, error)
	// 获取日志
	Log(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	// 加解密示例
	CipherExample(context.Context, *CipherExampleReq) (*CipherExampleResp, error)
	// mysql数据库示例
	MysqlExample(context.Context, *MysqlExampleReq) (*MysqlExampleResp, error)
	mustEmbedUnimplementedHelloServer()
}

// UnimplementedHelloServer must be embedded to have forward compatible implementations.
type UnimplementedHelloServer struct {
}

func (UnimplementedHelloServer) Say(context.Context, *SayReq) (*SayReq, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Say not implemented")
}
func (UnimplementedHelloServer) Call(context.Context, *SayReq) (*SayReq, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (UnimplementedHelloServer) Log(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Log not implemented")
}
func (UnimplementedHelloServer) CipherExample(context.Context, *CipherExampleReq) (*CipherExampleResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CipherExample not implemented")
}
func (UnimplementedHelloServer) MysqlExample(context.Context, *MysqlExampleReq) (*MysqlExampleResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MysqlExample not implemented")
}
func (UnimplementedHelloServer) mustEmbedUnimplementedHelloServer() {}

// UnsafeHelloServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HelloServer will
// result in compilation errors.
type UnsafeHelloServer interface {
	mustEmbedUnimplementedHelloServer()
}

func RegisterHelloServer(s grpc.ServiceRegistrar, srv HelloServer) {
	s.RegisterService(&Hello_ServiceDesc, srv)
}

func _Hello_Say_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SayReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServer).Say(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hello_Say_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServer).Say(ctx, req.(*SayReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hello_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SayReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hello_Call_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServer).Call(ctx, req.(*SayReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hello_Log_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServer).Log(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hello_Log_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServer).Log(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hello_CipherExample_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CipherExampleReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServer).CipherExample(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hello_CipherExample_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServer).CipherExample(ctx, req.(*CipherExampleReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hello_MysqlExample_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MysqlExampleReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServer).MysqlExample(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hello_MysqlExample_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServer).MysqlExample(ctx, req.(*MysqlExampleReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Hello_ServiceDesc is the grpc.ServiceDesc for Hello service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Hello_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.hello.Hello",
	HandlerType: (*HelloServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Say",
			Handler:    _Hello_Say_Handler,
		},
		{
			MethodName: "Call",
			Handler:    _Hello_Call_Handler,
		},
		{
			MethodName: "Log",
			Handler:    _Hello_Log_Handler,
		},
		{
			MethodName: "CipherExample",
			Handler:    _Hello_CipherExample_Handler,
		},
		{
			MethodName: "MysqlExample",
			Handler:    _Hello_MysqlExample_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hello/hello.proto",
}
