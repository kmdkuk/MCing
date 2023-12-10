package server

import (
	"github.com/james4k/rcon"
	"github.com/kmdkuk/mcing/pkg/proto"
)

// NewAgentService creates a new AgentServer
func NewAgentService(conn *rcon.RemoteConsole) proto.AgentServer {
	return agentService{
		conn: conn,
	}
}

type agentService struct {
	conn *rcon.RemoteConsole
	proto.UnimplementedAgentServer
}
