package server

import (
	"context"

	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
)

func (s agentService) Reload(ctx context.Context, req *proto.ReloadRequest) (*proto.ReloadResponse, error) {
	if err := rcon.Reload(s.conn); err != nil {
		return nil, err
	}
	return &proto.ReloadResponse{}, nil
}
