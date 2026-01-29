package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/kmdkuk/mcing/pkg/constants"
)

type MockConsole struct {
	WriteFunc func(cmd string) (int, error)
	ReadFunc  func() (string, int, error)
}

func (m *MockConsole) Write(cmd string) (int, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(cmd)
	}
	return 0, nil
}

func (m *MockConsole) Read() (string, int, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc()
	}
	return "", 0, nil
}

func TestWatch(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	dataDir := filepath.Join(tempDir, "data")
	err := os.MkdirAll(configDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(dataDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	// Initial config
	serverProps := filepath.Join(configDir, constants.ServerPropsName)
	err = os.WriteFile(serverProps, []byte("test=1"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	reloadCalled := false
	var mu sync.Mutex

	mock := &MockConsole{
		WriteFunc: func(cmd string) (int, error) {
			if cmd == "reload" {
				mu.Lock()
				reloadCalled = true
				mu.Unlock()
				return 1, nil
			}
			return 0, nil
		},
		ReadFunc: func() (string, int, error) {
			return "Reloaded", 1, nil
		},
	}

	cfg := Config{
		DataPath:   dataDir,
		ConfigPath: configDir,
	}

	ctx := t.Context()

	// Run Watch in background
	go func() {
		_ = Watch(ctx, mock, 100*time.Millisecond, cfg)
	}()

	// Wait a bit for initial sync (not strictly reloading, just reading)
	time.Sleep(200 * time.Millisecond)

	// Update config to trigger reload
	err = os.WriteFile(serverProps, []byte("test=2"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for reload
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	called := reloadCalled
	mu.Unlock()

	if !called {
		t.Error("expected Reload to be called after config change")
	}
}
