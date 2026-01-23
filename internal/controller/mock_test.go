package controller

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kmdkuk/mcing/internal/minecraft"
)

type mockManager struct {
	mu         sync.Mutex
	minecrafts map[string]struct{}
}

var _ minecraft.MinecraftManager = &mockManager{} //nolint:exhaustruct // interface check

func (m *mockManager) Update(key types.NamespacedName) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.minecrafts[key.String()] = struct{}{}
	return nil
}

func (m *mockManager) Stop(key types.NamespacedName) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.minecrafts, key.String())
}

func (m *mockManager) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
