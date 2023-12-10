package server

import (
	"context"
	"path"
	"strconv"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/kmdkuk/mcing/pkg/config"
	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
	"go.uber.org/zap"
)

func (s agentService) SyncWhitelist(ctx context.Context, req *proto.SyncWhitelistRequest) (*proto.SyncWhitelistResponse, error) {
	log := ctxzap.Extract(ctx).With(zap.String("func", "syncWhitelist"))
	log.Info("start sync white list")
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
	log.Info("current whitelist", zap.Strings("users", users))
	// add: Not present in users, but present in req.Users.
	addUsers := differenceSet(users, req.Users)
	if len(addUsers) > 0 {
		err := rcon.Whitelist(s.conn, "add", addUsers)
		if err != nil {
			return &proto.SyncWhitelistResponse{}, err
		}
	}
	// remove: Not present in req.Users, but present in users.
	removeUsers := differenceSet(req.Users, users)
	if len(removeUsers) > 0 {
		err := rcon.Whitelist(s.conn, "remove", removeUsers)
		if err != nil {
			return &proto.SyncWhitelistResponse{}, err
		}
	}
	log.Info("finish sync whitelist", zap.Strings("addUsers", addUsers), zap.Strings("removeUsers", removeUsers))
	return &proto.SyncWhitelistResponse{}, nil
}

func differenceSet(a, b []string) []string {
	exists := map[string]struct{}{}
	for _, v := range a {
		exists[v] = struct{}{}
	}

	differenceSet := make([]string, 0)
	for _, v := range b {
		if _, ok := exists[v]; !ok {
			differenceSet = append(differenceSet, v)
		}
	}
	return differenceSet
}
