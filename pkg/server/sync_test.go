package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
)

type MockConsole struct {
	WriteFunc func(cmd string) (int, error)
	ReadFunc  func() (string, int, error)
}

func (m *MockConsole) Write(cmd string) (int, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(cmd)
	}
	return 0, nil
}

func (m *MockConsole) Read() (string, int, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc()
	}
	return "", 0, nil
}

func TestSyncWhitelist(t *testing.T) {
	tempDir := t.TempDir()
	serverProps := filepath.Join(tempDir, constants.ServerPropsName)
	err := os.WriteFile(serverProps, []byte("white-list=true"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		req     *proto.SyncWhitelistRequest
		mock    rcon.Console
		wantErr bool
	}{
		{
			name: "enable whitelist",
			req: &proto.SyncWhitelistRequest{
				Enabled: true,
				Users:   []string{"user1"},
			},
			mock: &MockConsole{
				WriteFunc: func(_ string) (int, error) {
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					// ListWhitelist response
					return "There are no whitelisted players", 1, nil
				},
			},
			wantErr: false,
		},
	}

	logger := zap.NewNop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &agentService{
				UnimplementedAgentServer: proto.UnimplementedAgentServer{},
				logger:                   logger,
				conn:                     tt.mock,
				dataPath:                 tempDir,
			}
			_, err := s.SyncWhitelist(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncWhitelist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSyncOps(t *testing.T) {
	tempDir := t.TempDir()
	ops := filepath.Join(tempDir, constants.OpsName)
	err := os.WriteFile(ops, []byte("[]"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		req     *proto.SyncOpsRequest
		mock    rcon.Console
		wantErr bool
	}{
		{
			name: "add op",
			req: &proto.SyncOpsRequest{
				Users: []string{"user1"},
			},
			mock: &MockConsole{
				WriteFunc: func(cmd string) (int, error) {
					// op user1
					if cmd == "op user1" {
						return 1, nil
					}
					return 0, nil
				},
				ReadFunc: func() (string, int, error) {
					return "Made user1 a server operator", 1, nil
				},
			},
			wantErr: false,
		},
	}

	logger := zap.NewNop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &agentService{
				UnimplementedAgentServer: proto.UnimplementedAgentServer{},
				logger:                   logger,
				conn:                     tt.mock,
				dataPath:                 tempDir,
			}
			_, err := s.SyncOps(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncOps() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
