package minecraft

import (
	"context"

	"google.golang.org/grpc"

	"github.com/kmdkuk/mcing/pkg/agent"
	"github.com/kmdkuk/mcing/pkg/proto"
)

type mockAgentConn struct {
	reloadFunc        func(ctx context.Context, in *proto.ReloadRequest, opts ...grpc.CallOption) (*proto.ReloadResponse, error)
	syncWhitelistFunc func(ctx context.Context, in *proto.SyncWhitelistRequest, opts ...grpc.CallOption) (*proto.SyncWhitelistResponse, error)
	syncOpsFunc       func(ctx context.Context, in *proto.SyncOpsRequest, opts ...grpc.CallOption) (*proto.SyncOpsResponse, error)
}

func (m *mockAgentConn) Reload(
	ctx context.Context,
	in *proto.ReloadRequest,
	opts ...grpc.CallOption,
) (*proto.ReloadResponse, error) {
	if m.reloadFunc != nil {
		return m.reloadFunc(ctx, in, opts...)
	}
	return &proto.ReloadResponse{}, nil
}

func (m *mockAgentConn) SyncWhitelist(
	ctx context.Context,
	in *proto.SyncWhitelistRequest,
	opts ...grpc.CallOption,
) (*proto.SyncWhitelistResponse, error) {
	if m.syncWhitelistFunc != nil {
		return m.syncWhitelistFunc(ctx, in, opts...)
	}
	return &proto.SyncWhitelistResponse{}, nil
}

func (m *mockAgentConn) SyncOps(
	ctx context.Context,
	in *proto.SyncOpsRequest,
	opts ...grpc.CallOption,
) (*proto.SyncOpsResponse, error) {
	if m.syncOpsFunc != nil {
		return m.syncOpsFunc(ctx, in, opts...)
	}
	return &proto.SyncOpsResponse{}, nil
}

func (m *mockAgentConn) Close() error {
	return nil
}

var _ agent.Conn = &mockAgentConn{} //nolint:exhaustruct // interface check
