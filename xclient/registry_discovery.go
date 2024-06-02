package xclient

import (
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type WheelRegistryDiscovery struct {
	*MultiServersDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewWheelRegistryDiscovery(registry string, timeout time.Duration) *WheelRegistryDiscovery {
	if timeout == -0 {
		timeout = defaultUpdateTimeout
	}
	return &WheelRegistryDiscovery{
		MultiServersDiscovery: &MultiServersDiscovery{
			servers: make([]string, 0),
			r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		},
		registry: registry,
		timeout:  timeout,
	}
}

func (d *WheelRegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *WheelRegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("refreshing registry:", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("refresh registry failed:", err)
		return err
	}
	str := resp.Header.Get("X-Wheelrpc-Servers")
	servers := strings.Split(str, ",")
	d.servers = servers
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, server)
		}
	}
	d.lastUpdate = time.Now()
	return err
}

func (d *WheelRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServersDiscovery.Get(mode)
}

func (d *WheelRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServersDiscovery.GetAll()
}
