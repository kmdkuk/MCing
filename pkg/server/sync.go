package server

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"strconv"

	"go.uber.org/zap"

	"github.com/kmdkuk/mcing/pkg/config"
	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
)

func (s agentService) SyncWhitelist(
	_ context.Context,
	req *proto.SyncWhitelistRequest,
) (*proto.SyncWhitelistResponse, error) {
	log := s.logger.With(zap.String("func", "syncWhitelist"))
	log.Info("start sync white list")
	// parse /data/server.peroperties using config.ParseServerProps
	props, err := config.ParseServerPropsFromPath(path.Join(constants.DataPath, constants.ServerPropsName))
	if err != nil {
		return &proto.SyncWhitelistResponse{}, err
	}
	enabled, err := strconv.ParseBool(props[constants.WhitelistProps])
	if err != nil {
		return &proto.SyncWhitelistResponse{}, err
	}
	if enabled != req.GetEnabled() {
		err = rcon.WhitelistSwitch(s.conn, req.GetEnabled())
		if err != nil {
			return &proto.SyncWhitelistResponse{}, err
		}
	}
	if !req.GetEnabled() {
		return &proto.SyncWhitelistResponse{}, nil
	}
	users, err := rcon.ListWhitelist(s.conn)
	if err != nil {
		return &proto.SyncWhitelistResponse{}, err
	}
	log.Info("current whitelist", zap.Strings("users", users))
	// add: Not present in users, but present in req.Users.
	addUsers := differenceSet(users, req.GetUsers())
	if len(addUsers) > 0 {
		if err := rcon.Whitelist(s.conn, "add", addUsers); err != nil {
			return &proto.SyncWhitelistResponse{}, err
		}
	}
	// remove: Not present in req.Users, but present in users.
	removeUsers := differenceSet(req.GetUsers(), users)
	if len(removeUsers) > 0 {
		err := rcon.Whitelist(s.conn, "remove", removeUsers)
		if err != nil {
			return &proto.SyncWhitelistResponse{}, err
		}
	}
	log.Info("finish sync whitelist", zap.Strings("addUsers", addUsers), zap.Strings("removeUsers", removeUsers))
	return &proto.SyncWhitelistResponse{}, nil
}

type opsJSON struct {
	UUID                string `json:"uuid"`
	Name                string `json:"name"`
	Level               int    `json:"level"`
	BypassesPlayerLimit bool   `json:"bypassesPlayerLimit"`
}

func (s agentService) SyncOps(_ context.Context, req *proto.SyncOpsRequest) (*proto.SyncOpsResponse, error) {
	log := s.logger.With(zap.String("func", "syncOps"))
	log.Info("start sync ops")
	raw, err := os.ReadFile(path.Join(constants.DataPath, constants.OpsName))
	if err != nil {
		return &proto.SyncOpsResponse{}, err
	}
	var ops []opsJSON
	if err := json.Unmarshal(raw, &ops); err != nil {
		return &proto.SyncOpsResponse{}, err
	}
	users := make([]string, 0)
	for _, v := range ops {
		users = append(users, v.Name)
	}

	addUsers := differenceSet(users, req.GetUsers())
	if len(addUsers) > 0 {
		err := rcon.Op(s.conn, addUsers)
		if err != nil {
			return &proto.SyncOpsResponse{}, err
		}
	}
	// remove: Not present in req.Users, but present in users.
	removeUsers := differenceSet(req.GetUsers(), users)
	if len(removeUsers) > 0 {
		err := rcon.Deop(s.conn, removeUsers)
		if err != nil {
			return &proto.SyncOpsResponse{}, err
		}
	}
	log.Info("finish sync Ops", zap.Strings("addUsers", addUsers), zap.Strings("removeUsers", removeUsers))
	return &proto.SyncOpsResponse{}, nil
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
