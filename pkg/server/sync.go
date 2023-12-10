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
		return nil, err
	}
	enabled, err := strconv.ParseBool(props[constants.Whitelist])
	if err != nil {
		return nil, err
	}
	if enabled != req.Enabled {
		rcon.WhitelistSwitch(req.Enabled)
	}
	if !req.Enabled {
		return nil, nil
	}
	users := rcon.Whitelistlist()
	fmt.Println(users)
	return nil, nil
}
