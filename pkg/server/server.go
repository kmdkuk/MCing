package server

import (
	"github.com/james4k/rcon"
	"github.com/kmdkuk/mcing/pkg/proto"
	"go.uber.org/zap"
)

// NewAgentService creates a new AgentServer
func NewAgentService(logger *zap.Logger, conn *rcon.RemoteConsole) proto.AgentServer {
	return agentService{
		logger: logger.With(zap.String("service", "mcing-agent")),
		conn:   conn,
	}
}

type agentService struct {
	logger *zap.Logger
	conn   *rcon.RemoteConsole
	proto.UnimplementedAgentServer
}
