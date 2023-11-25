package server

import "github.com/kmdkuk/mcing/pkg/proto"

// NewAgentService creates a new AgentServer
func NewAgentService(agent *Agent) proto.AgentServer {
	return agentService{agent: agent}
}

type agentService struct {
	agent *Agent
	proto.UnimplementedAgentServer
}

func New() *Agent {
	return &Agent{}
}

type Agent struct {
}
