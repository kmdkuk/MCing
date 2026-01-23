package server

import (
	"github.com/james4k/rcon"
	"go.uber.org/zap"

	"github.com/kmdkuk/mcing/pkg/proto"
)

// NewAgentService creates a new AgentServer.
func NewAgentService(logger *zap.Logger, conn *rcon.RemoteConsole) proto.AgentServer {
	return agentService{ //nolint:exhaustruct // unimplemented embedded struct
		logger: logger.With(zap.String("service", "mcing-agent")),
		conn:   conn,
	}
}

type agentService struct {
	proto.UnimplementedAgentServer

	logger *zap.Logger
	conn   *rcon.RemoteConsole
}
