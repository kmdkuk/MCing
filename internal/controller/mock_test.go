package controller

import (
	"sync"

	"github.com/kmdkuk/mcing/internal/minecraft"
	"k8s.io/apimachinery/pkg/types"
)

type mockManager struct {
	mu         sync.Mutex
	minecrafts map[string]struct{}
}

var _ minecraft.MinecraftManager = &mockManager{}

func (m *mockManager) Update(key types.NamespacedName) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.minecrafts[key.String()] = struct{}{}
}

func (m *mockManager) Stop(key types.NamespacedName) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.minecrafts, key.String())
}

func (m *mockManager) StopAll() {}
