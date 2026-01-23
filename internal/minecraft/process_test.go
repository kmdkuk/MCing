package minecraft

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/proto"
)

func Test_managerProcess_sync(t *testing.T) {
	type args struct {
		mc *mcingv1alpha1.Minecraft
	}
	tests := []struct {
		name              string
		args              args
		syncWhitelistFunc func(ctx context.Context, in *proto.SyncWhitelistRequest, opts ...grpc.CallOption) (*proto.SyncWhitelistResponse, error)
		syncOpsFunc       func(ctx context.Context, in *proto.SyncOpsRequest, opts ...grpc.CallOption) (*proto.SyncOpsResponse, error)
		wantErr           bool
	}{
		{
			name: "success",
			args: args{
				mc: &mcingv1alpha1.Minecraft{
					Spec: mcingv1alpha1.MinecraftSpec{
						Whitelist: mcingv1alpha1.Whitelist{
							Enabled: true,
							Users:   []string{"user1"},
						},
						Ops: mcingv1alpha1.Ops{
							Users: []string{"op1"},
						},
					},
				},
			},
			syncWhitelistFunc: func(_ context.Context, in *proto.SyncWhitelistRequest, _ ...grpc.CallOption) (*proto.SyncWhitelistResponse, error) {
				if !in.GetEnabled() {
					t.Errorf("expected whitelist enabled, got false")
				}
				if len(in.GetUsers()) != 1 || in.GetUsers()[0] != "user1" {
					t.Errorf("expected whitelist users [user1], got %v", in.GetUsers())
				}
				return &proto.SyncWhitelistResponse{}, nil
			},
			syncOpsFunc: func(_ context.Context, in *proto.SyncOpsRequest, _ ...grpc.CallOption) (*proto.SyncOpsResponse, error) {
				if len(in.GetUsers()) != 1 || in.GetUsers()[0] != "op1" {
					t.Errorf("expected ops users [op1], got %v", in.GetUsers())
				}
				return &proto.SyncOpsResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "whitelist error",
			args: args{
				mc: &mcingv1alpha1.Minecraft{},
			},
			syncWhitelistFunc: func(_ context.Context, _ *proto.SyncWhitelistRequest, _ ...grpc.CallOption) (*proto.SyncWhitelistResponse, error) {
				return nil, errors.New("whitelist error")
			},
			wantErr: true,
		},
		{
			name: "ops error",
			args: args{
				mc: &mcingv1alpha1.Minecraft{},
			},
			syncOpsFunc: func(_ context.Context, _ *proto.SyncOpsRequest, _ ...grpc.CallOption) (*proto.SyncOpsResponse, error) {
				return nil, errors.New("ops error")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &managerProcess{ //nolint:exhaustruct // internal struct
				log: logr.Discard(),
			}
			agent := &mockAgentConn{ //nolint:exhaustruct // internal struct
				syncWhitelistFunc: tt.syncWhitelistFunc,
				syncOpsFunc:       tt.syncOpsFunc,
			}
			if err := p.sync(context.Background(), tt.args.mc, agent); (err != nil) != tt.wantErr {
				t.Errorf("managerProcess.sync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
