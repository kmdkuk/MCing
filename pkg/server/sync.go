package server

import (
	"context"
	"fmt"
	"path"
	"strconv"

	"github.com/kmdkuk/mcing/pkg/config"
	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
)

func (s agentService) SyncWhitelist(ctx context.Context, req *proto.SyncWhitelistRequest) (*proto.SyncWhitelistResponse, error) {
	// parse /data/server.peroperties using config.ParseServerProps
	props, err := config.ParseServerProps(path.Join(constants.DataPath, constants.ServerPropsName))
	if err != nil {
		return &proto.SyncWhitelistResponse{}, err
	}
	enabled, err := strconv.ParseBool(props[constants.WhitelistProps])
	if err != nil {
		return &proto.SyncWhitelistResponse{}, err
	}
	if enabled != req.Enabled {
		rcon.WhitelistSwitch(s.conn, req.Enabled)
	}
	if !req.Enabled {
		return &proto.SyncWhitelistResponse{}, nil
	}
	users, err := rcon.Whitelistlist(s.conn)
	if err != nil {
		return &proto.SyncWhitelistResponse{}, err
	}
	fmt.Println(users)
	return &proto.SyncWhitelistResponse{}, nil
}
