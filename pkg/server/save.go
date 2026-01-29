package server

import (
	"context"

	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
)

func (s agentService) SaveOff(_ context.Context, _ *proto.SaveOffRequest) (*proto.SaveOffResponse, error) {
	err := rcon.SaveOff(s.conn)
	if err != nil {
		return nil, err
	}
	return &proto.SaveOffResponse{}, nil
}

func (s agentService) SaveAll(_ context.Context, _ *proto.SaveAllRequest) (*proto.SaveAllResponse, error) {
	err := rcon.SaveAllFlush(s.conn)
	if err != nil {
		return nil, err
	}
	return &proto.SaveAllResponse{}, nil
}

func (s agentService) SaveOn(_ context.Context, _ *proto.SaveOnRequest) (*proto.SaveOnResponse, error) {
	err := rcon.SaveOn(s.conn)
	if err != nil {
		return nil, err
	}
	return &proto.SaveOnResponse{}, nil
}
