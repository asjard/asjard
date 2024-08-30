// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.27.0
// source: request.proto

package requestpb

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
	DefaultHandlers_Favicon_FullMethodName = "/asjard.api.DefaultHandlers/Favicon"
)

// DefaultHandlersClient is the client API for DefaultHandlers service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DefaultHandlersClient interface {
	// Return favicon.ico
	Favicon(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type defaultHandlersClient struct {
	cc grpc.ClientConnInterface
}

func NewDefaultHandlersClient(cc grpc.ClientConnInterface) DefaultHandlersClient {
	return &defaultHandlersClient{cc}
}

func (c *defaultHandlersClient) Favicon(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, DefaultHandlers_Favicon_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DefaultHandlersServer is the server API for DefaultHandlers service.
// All implementations must embed UnimplementedDefaultHandlersServer
// for forward compatibility
type DefaultHandlersServer interface {
	// Return favicon.ico
	Favicon(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedDefaultHandlersServer()
}

// UnimplementedDefaultHandlersServer must be embedded to have forward compatible implementations.
type UnimplementedDefaultHandlersServer struct {
}

func (UnimplementedDefaultHandlersServer) Favicon(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Favicon not implemented")
}
func (UnimplementedDefaultHandlersServer) mustEmbedUnimplementedDefaultHandlersServer() {}

// UnsafeDefaultHandlersServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DefaultHandlersServer will
// result in compilation errors.
type UnsafeDefaultHandlersServer interface {
	mustEmbedUnimplementedDefaultHandlersServer()
}

func RegisterDefaultHandlersServer(s grpc.ServiceRegistrar, srv DefaultHandlersServer) {
	s.RegisterService(&DefaultHandlers_ServiceDesc, srv)
}

func _DefaultHandlers_Favicon_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DefaultHandlersServer).Favicon(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DefaultHandlers_Favicon_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DefaultHandlersServer).Favicon(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// DefaultHandlers_ServiceDesc is the grpc.ServiceDesc for DefaultHandlers service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DefaultHandlers_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "asjard.api.DefaultHandlers",
	HandlerType: (*DefaultHandlersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Favicon",
			Handler:    _DefaultHandlers_Favicon_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "request.proto",
}
