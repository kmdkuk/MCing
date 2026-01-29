package agent

import (
	"context"
	"testing"
)

func TestNewFactory(t *testing.T) {
	f := NewFactory()
	if f == nil {
		t.Error("NewFactory() returned nil")
	}
}

func TestDefaultAgentFactory_New(t *testing.T) {
	f := NewFactory()
	// Using a dummy address; grpc.NewClient is non-blocking and shouldn't fail immediately
	// merely because the target is unreachable, unless blocking options are used.
	// The implementation uses insecure creds and standard params.
	conn, err := f.New(context.Background(), "127.0.0.1")
	if err != nil {
		t.Errorf("defaultAgentFactory.New() error = %v", err)
	}
	if conn == nil {
		t.Error("defaultAgentFactory.New() returned nil connection")
	}
}
