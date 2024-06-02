package xclient

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

type MultiServersDiscovery struct {
	r       *rand.Rand
	mu      sync.RWMutex
	servers []string
	index   int
}

var (
	_                       Discovery = (*MultiServersDiscovery)(nil)
	ErrNoAvailableServer              = errors.New("no available servers")
	ErrNotSupportSelectMode           = errors.New("not support select mode")
)

func (m *MultiServersDiscovery) Refresh() error {
	return nil
}

func (m *MultiServersDiscovery) Update(servers []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers = servers
	return nil
}

func (m *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.servers) == 0 {
		return "", ErrNoAvailableServer
	}
	switch mode {
	case RandomSelect:
		return m.servers[m.r.Intn(len(m.servers))], nil
	case RoundRobinSelect:
		s := m.servers[m.index]
		m.index = (m.index + 1) % len(m.servers)
		return s, nil
	default:
		return "", ErrNotSupportSelectMode
	}
}

func (m *MultiServersDiscovery) GetAll() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	servers := make([]string, len(m.servers))
	copy(servers, m.servers)
	return servers, nil
}

func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().Unix())),
	}
	d.index = d.r.Intn(len(servers))
	return d
}
