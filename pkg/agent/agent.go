package agent

import (
	"context"
	"io"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/kmdkuk/mcing/pkg/constants"
	agent "github.com/kmdkuk/mcing/pkg/proto"
)

const (
	keepaliveTime     = 1 * time.Minute
	backoffBaseDelay  = 1.0 * time.Second
	backoffMultiplier = 1.6
	backoffJitter     = 0.2
	backoffMaxDelay   = 30 * time.Second
	minConnectTimeout = 20 * time.Second
)

// Conn represents a gRPC connection to a mcing-agent.
type Conn interface {
	agent.AgentClient
	io.Closer
}

type agentConn struct {
	agent.AgentClient
	*grpc.ClientConn
}

var _ Conn = agentConn{} //nolint:exhaustruct // interface check

// Factory represents the interface of a factory to create Conn.
type Factory interface {
	New(ctx context.Context, podIP string) (Conn, error)
}

// NewFactory returns a new Factory.
func NewFactory() Factory {
	return defaultAgentFactory{}
}

type defaultAgentFactory struct{}

var _ Factory = defaultAgentFactory{}

func (f defaultAgentFactory) New(_ context.Context, podIP string) (Conn, error) {
	addr := net.JoinHostPort(podIP, strconv.Itoa(int(constants.AgentPort)))
	kp := keepalive.ClientParameters{
		Time:                keepaliveTime,
		Timeout:             0,
		PermitWithoutStream: false,
	}
	cp := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  backoffBaseDelay,
			Multiplier: backoffMultiplier,
			Jitter:     backoffJitter,
			MaxDelay:   backoffMaxDelay,
		},
		MinConnectTimeout: minConnectTimeout,
	}
	conn, err := grpc.NewClient(addr,
		grpc.WithKeepaliveParams(kp),
		grpc.WithConnectParams(cp),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return agentConn{}, err
	}
	return agentConn{
		AgentClient: agent.NewAgentClient(conn),
		ClientConn:  conn,
	}, nil
}
