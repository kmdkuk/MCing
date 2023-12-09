package minecraft

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type MinecraftManager interface {
	Update(types.NamespacedName)
	Stop(types.NamespacedName)
	StopAll()
}

func NewManager(interval time.Duration, m manager.Manager, log logr.Logger) MinecraftManager {
	return &minecraftManager{
		k8sclient: m.GetClient(),
		interval:  interval,
		log:       log,
		processes: make(map[string]*managerProcess),
	}
}

type minecraftManager struct {
	k8sclient client.Client
	interval  time.Duration
	log       logr.Logger

	mu        sync.Mutex
	processes map[string]*managerProcess

	wg sync.WaitGroup
}

func (m *minecraftManager) Update(name types.NamespacedName) {
	m.update(name)
}

func (m *minecraftManager) update(name types.NamespacedName) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name.String()
	p, ok := m.processes[key]
	if ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	log := m.log.WithName(key)
	p = newManagerProcess(m.k8sclient, name, log, cancel)
	m.wg.Add(1)
	go func() {
		p.Start(ctx, m.interval)
		m.wg.Done()
	}()
	m.processes[key] = p
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

func (m *minecraftManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.processes {
		p.Cancel()
	}
	m.processes = nil

	m.wg.Wait()
}
