package minecraft

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/proto"
	"google.golang.org/grpc"
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
			syncWhitelistFunc: func(ctx context.Context, in *proto.SyncWhitelistRequest, opts ...grpc.CallOption) (*proto.SyncWhitelistResponse, error) {
				if !in.Enabled {
					t.Errorf("expected whitelist enabled, got false")
				}
				if len(in.Users) != 1 || in.Users[0] != "user1" {
					t.Errorf("expected whitelist users [user1], got %v", in.Users)
				}
				return &proto.SyncWhitelistResponse{}, nil
			},
			syncOpsFunc: func(ctx context.Context, in *proto.SyncOpsRequest, opts ...grpc.CallOption) (*proto.SyncOpsResponse, error) {
				if len(in.Users) != 1 || in.Users[0] != "op1" {
					t.Errorf("expected ops users [op1], got %v", in.Users)
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
			syncWhitelistFunc: func(ctx context.Context, in *proto.SyncWhitelistRequest, opts ...grpc.CallOption) (*proto.SyncWhitelistResponse, error) {
				return nil, errors.New("whitelist error")
			},
			wantErr: true,
		},
		{
			name: "ops error",
			args: args{
				mc: &mcingv1alpha1.Minecraft{},
			},
			syncOpsFunc: func(ctx context.Context, in *proto.SyncOpsRequest, opts ...grpc.CallOption) (*proto.SyncOpsResponse, error) {
				return nil, errors.New("ops error")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &managerProcess{
				log: logr.Discard(),
			}
			agent := &mockAgentConn{
				syncWhitelistFunc: tt.syncWhitelistFunc,
				syncOpsFunc:       tt.syncOpsFunc,
			}
			if err := p.sync(context.Background(), tt.args.mc, agent); (err != nil) != tt.wantErr {
				t.Errorf("managerProcess.sync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
