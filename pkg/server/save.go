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

func (s agentService) SaveAllFlush(
	_ context.Context,
	_ *proto.SaveAllFlushRequest,
) (*proto.SaveAllFlushResponse, error) {
	err := rcon.SaveAllFlush(s.conn)
	if err != nil {
		return nil, err
	}
	return &proto.SaveAllFlushResponse{}, nil
}

func (s agentService) SaveOn(_ context.Context, _ *proto.SaveOnRequest) (*proto.SaveOnResponse, error) {
	err := rcon.SaveOn(s.conn)
	if err != nil {
		return nil, err
	}
	return &proto.SaveOnResponse{}, nil
}
