// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.1
// source: locket.proto

package models

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Locket_Lock_FullMethodName     = "/models.Locket/Lock"
	Locket_Fetch_FullMethodName    = "/models.Locket/Fetch"
	Locket_Release_FullMethodName  = "/models.Locket/Release"
	Locket_FetchAll_FullMethodName = "/models.Locket/FetchAll"
)

// LocketClient is the client API for Locket service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LocketClient interface {
	Lock(ctx context.Context, in *LockRequest, opts ...grpc.CallOption) (*LockResponse, error)
	Fetch(ctx context.Context, in *FetchRequest, opts ...grpc.CallOption) (*FetchResponse, error)
	Release(ctx context.Context, in *ReleaseRequest, opts ...grpc.CallOption) (*ReleaseResponse, error)
	FetchAll(ctx context.Context, in *FetchAllRequest, opts ...grpc.CallOption) (*FetchAllResponse, error)
}

type locketClient struct {
	cc grpc.ClientConnInterface
}

func NewLocketClient(cc grpc.ClientConnInterface) LocketClient {
	return &locketClient{cc}
}

func (c *locketClient) Lock(ctx context.Context, in *LockRequest, opts ...grpc.CallOption) (*LockResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LockResponse)
	err := c.cc.Invoke(ctx, Locket_Lock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locketClient) Fetch(ctx context.Context, in *FetchRequest, opts ...grpc.CallOption) (*FetchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(FetchResponse)
	err := c.cc.Invoke(ctx, Locket_Fetch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locketClient) Release(ctx context.Context, in *ReleaseRequest, opts ...grpc.CallOption) (*ReleaseResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReleaseResponse)
	err := c.cc.Invoke(ctx, Locket_Release_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locketClient) FetchAll(ctx context.Context, in *FetchAllRequest, opts ...grpc.CallOption) (*FetchAllResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(FetchAllResponse)
	err := c.cc.Invoke(ctx, Locket_FetchAll_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LocketServer is the server API for Locket service.
// All implementations must embed UnimplementedLocketServer
// for forward compatibility.
type LocketServer interface {
	Lock(context.Context, *LockRequest) (*LockResponse, error)
	Fetch(context.Context, *FetchRequest) (*FetchResponse, error)
	Release(context.Context, *ReleaseRequest) (*ReleaseResponse, error)
	FetchAll(context.Context, *FetchAllRequest) (*FetchAllResponse, error)
	mustEmbedUnimplementedLocketServer()
}

// UnimplementedLocketServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedLocketServer struct{}

func (UnimplementedLocketServer) Lock(context.Context, *LockRequest) (*LockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Lock not implemented")
}
func (UnimplementedLocketServer) Fetch(context.Context, *FetchRequest) (*FetchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Fetch not implemented")
}
func (UnimplementedLocketServer) Release(context.Context, *ReleaseRequest) (*ReleaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Release not implemented")
}
func (UnimplementedLocketServer) FetchAll(context.Context, *FetchAllRequest) (*FetchAllResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchAll not implemented")
}
func (UnimplementedLocketServer) mustEmbedUnimplementedLocketServer() {}
func (UnimplementedLocketServer) testEmbeddedByValue()                {}

// UnsafeLocketServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LocketServer will
// result in compilation errors.
type UnsafeLocketServer interface {
	mustEmbedUnimplementedLocketServer()
}

func RegisterLocketServer(s grpc.ServiceRegistrar, srv LocketServer) {
	// If the following call pancis, it indicates UnimplementedLocketServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Locket_ServiceDesc, srv)
}

func _Locket_Lock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocketServer).Lock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Locket_Lock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocketServer).Lock(ctx, req.(*LockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Locket_Fetch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocketServer).Fetch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Locket_Fetch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocketServer).Fetch(ctx, req.(*FetchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Locket_Release_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReleaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocketServer).Release(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Locket_Release_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocketServer).Release(ctx, req.(*ReleaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Locket_FetchAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchAllRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocketServer).FetchAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Locket_FetchAll_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocketServer).FetchAll(ctx, req.(*FetchAllRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Locket_ServiceDesc is the grpc.ServiceDesc for Locket service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Locket_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "models.Locket",
	HandlerType: (*LocketServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Lock",
			Handler:    _Locket_Lock_Handler,
		},
		{
			MethodName: "Fetch",
			Handler:    _Locket_Fetch_Handler,
		},
		{
			MethodName: "Release",
			Handler:    _Locket_Release_Handler,
		},
		{
			MethodName: "FetchAll",
			Handler:    _Locket_FetchAll_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "locket.proto",
}
