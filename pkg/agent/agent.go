package agent

import (
	"context"
	"io"
	"net"
	"strconv"
	"time"

	agent "github.com/kmdkuk/mcing-agent/proto"
	"github.com/kmdkuk/mcing/pkg/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// AgentConn represents a gRPC connection to a mcing-agent
type AgentConn interface {
	agent.AgentClient
	io.Closer
}

type agentConn struct {
	agent.AgentClient
	*grpc.ClientConn
}

var _ AgentConn = agentConn{}

// AgentFactory represents the interface of a factory to create AgentConn
type AgentFactory interface {
	New(ctx context.Context, podIP string) (AgentConn, error)
}

// NewAgentFactory returns a new AgentFactory.
func NewAgentFactory() AgentFactory {
	return defaultAgentFactory{}
}

type defaultAgentFactory struct {
}

var _ AgentFactory = defaultAgentFactory{}

func (f defaultAgentFactory) New(ctx context.Context, podIP string) (AgentConn, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	addr := net.JoinHostPort(podIP, strconv.Itoa(int(constants.AgentPort)))
	kp := keepalive.ClientParameters{
		Time: 1 * time.Minute,
	}
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(kp),
		grpc.WithInsecure(),
	)
	if err != nil {
		return agentConn{}, err
	}
	return agentConn{
		AgentClient: agent.NewAgentClient(conn),
		ClientConn:  conn,
	}, nil
}
