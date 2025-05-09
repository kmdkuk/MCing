// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.2
// source: pkg/proto/agentrpc.proto

package proto

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
	Agent_Reload_FullMethodName        = "/mcing.Agent/Reload"
	Agent_SyncWhitelist_FullMethodName = "/mcing.Agent/SyncWhitelist"
	Agent_SyncOps_FullMethodName       = "/mcing.Agent/SyncOps"
)

// AgentClient is the client API for Agent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// *
// Agent provides services for MCing.
type AgentClient interface {
	Reload(ctx context.Context, in *ReloadRequest, opts ...grpc.CallOption) (*ReloadResponse, error)
	SyncWhitelist(ctx context.Context, in *SyncWhitelistRequest, opts ...grpc.CallOption) (*SyncWhitelistResponse, error)
	SyncOps(ctx context.Context, in *SyncOpsRequest, opts ...grpc.CallOption) (*SyncOpsResponse, error)
}

type agentClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentClient(cc grpc.ClientConnInterface) AgentClient {
	return &agentClient{cc}
}

func (c *agentClient) Reload(ctx context.Context, in *ReloadRequest, opts ...grpc.CallOption) (*ReloadResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReloadResponse)
	err := c.cc.Invoke(ctx, Agent_Reload_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) SyncWhitelist(ctx context.Context, in *SyncWhitelistRequest, opts ...grpc.CallOption) (*SyncWhitelistResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SyncWhitelistResponse)
	err := c.cc.Invoke(ctx, Agent_SyncWhitelist_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) SyncOps(ctx context.Context, in *SyncOpsRequest, opts ...grpc.CallOption) (*SyncOpsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SyncOpsResponse)
	err := c.cc.Invoke(ctx, Agent_SyncOps_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AgentServer is the server API for Agent service.
// All implementations must embed UnimplementedAgentServer
// for forward compatibility.
//
// *
// Agent provides services for MCing.
type AgentServer interface {
	Reload(context.Context, *ReloadRequest) (*ReloadResponse, error)
	SyncWhitelist(context.Context, *SyncWhitelistRequest) (*SyncWhitelistResponse, error)
	SyncOps(context.Context, *SyncOpsRequest) (*SyncOpsResponse, error)
	mustEmbedUnimplementedAgentServer()
}

// UnimplementedAgentServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAgentServer struct{}

func (UnimplementedAgentServer) Reload(context.Context, *ReloadRequest) (*ReloadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reload not implemented")
}
func (UnimplementedAgentServer) SyncWhitelist(context.Context, *SyncWhitelistRequest) (*SyncWhitelistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncWhitelist not implemented")
}
func (UnimplementedAgentServer) SyncOps(context.Context, *SyncOpsRequest) (*SyncOpsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncOps not implemented")
}
func (UnimplementedAgentServer) mustEmbedUnimplementedAgentServer() {}
func (UnimplementedAgentServer) testEmbeddedByValue()               {}

// UnsafeAgentServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AgentServer will
// result in compilation errors.
type UnsafeAgentServer interface {
	mustEmbedUnimplementedAgentServer()
}

func RegisterAgentServer(s grpc.ServiceRegistrar, srv AgentServer) {
	// If the following call pancis, it indicates UnimplementedAgentServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Agent_ServiceDesc, srv)
}

func _Agent_Reload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReloadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).Reload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_Reload_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).Reload(ctx, req.(*ReloadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_SyncWhitelist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SyncWhitelistRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).SyncWhitelist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_SyncWhitelist_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).SyncWhitelist(ctx, req.(*SyncWhitelistRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_SyncOps_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SyncOpsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).SyncOps(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_SyncOps_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).SyncOps(ctx, req.(*SyncOpsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Agent_ServiceDesc is the grpc.ServiceDesc for Agent service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Agent_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mcing.Agent",
	HandlerType: (*AgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Reload",
			Handler:    _Agent_Reload_Handler,
		},
		{
			MethodName: "SyncWhitelist",
			Handler:    _Agent_SyncWhitelist_Handler,
		},
		{
			MethodName: "SyncOps",
			Handler:    _Agent_SyncOps_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/proto/agentrpc.proto",
}
