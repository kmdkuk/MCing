package server

import (
	"context"

	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
)

func (s agentService) Reload(ctx context.Context, req *proto.ReloadRequest) (*proto.ReloadResponse, error) {
	if err := s.agent.Reload(ctx, req); err != nil {
		return nil, err
	}
	return &proto.ReloadResponse{}, nil
}

func (a *Agent) Reload(ctx context.Context, req *proto.ReloadRequest) error {
	rcon.Reload()
	return nil
}
