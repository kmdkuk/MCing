package server

import "github.com/kmdkuk/mcing/pkg/proto"

// NewAgentService creates a new AgentServer
func NewAgentService() proto.AgentServer {
	return agentService{}
}

type agentService struct {
	proto.UnimplementedAgentServer
}
