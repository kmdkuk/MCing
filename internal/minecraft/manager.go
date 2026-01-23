package minecraft

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kmdkuk/mcing/pkg/agent"
)

// MinecraftManager manages the lifecycle of Minecraft server.
type MinecraftManager interface { //nolint:revive // MinecraftManager is exported identifier
	Update(types.NamespacedName) error
	Stop(types.NamespacedName)
	Start(context.Context) error
}

// NewManager creates a new MinecraftManager.
func NewManager(af agent.Factory, interval time.Duration, m manager.Manager, log logr.Logger) MinecraftManager {
	return &minecraftManager{ //nolint:exhaustruct // internal struct initialized efficiently
		af:        af,
		k8sclient: m.GetClient(),
		interval:  interval,
		log:       log,
		processes: make(map[string]*managerProcess),
	}
}

type minecraftManager struct {
	af        agent.Factory
	k8sclient client.Client
	interval  time.Duration
	log       logr.Logger

	mu        sync.Mutex
	processes map[string]*managerProcess

	wg sync.WaitGroup
}

func (m *minecraftManager) Start(ctx context.Context) error {
	<-ctx.Done()
	m.stopAll()
	return nil
}

func (m *minecraftManager) Update(name types.NamespacedName) error {
	return m.update(name)
}

func (m *minecraftManager) update(name types.NamespacedName) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name.String()
	p, ok := m.processes[key]
	if ok {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	log := m.log.WithName(key)
	p = newManagerProcess(m.af, m.k8sclient, name, log, cancel)
	m.wg.Go(func() {
		p.Start(ctx, m.interval)
	})
	m.processes[key] = p
	return nil
}

func (m *minecraftManager) Stop(name types.NamespacedName) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name.String()
	p, ok := m.processes[key]
	if ok {
		p.Cancel()
		delete(m.processes, key)
	}
}

func (m *minecraftManager) stopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.processes {
		p.Cancel()
	}
	m.processes = nil

	m.wg.Wait()
}
