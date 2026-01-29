package server

import (
	"go.uber.org/zap"

	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
)

// NewAgentService creates a new AgentServer.
// NewAgentService creates a new AgentServer.
func NewAgentService(logger *zap.Logger, conn rcon.Console) proto.AgentServer {
	return agentService{ //nolint:exhaustruct // unimplemented embedded struct
		logger:   logger.With(zap.String("service", "mcing-agent")),
		conn:     conn,
		dataPath: constants.DataPath,
	}
}

type agentService struct {
	proto.UnimplementedAgentServer

	logger   *zap.Logger
	conn     rcon.Console
	dataPath string
}
