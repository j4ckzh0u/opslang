package ssh

import (
	"fmt"
	"sync"
)

// Pool manages reusable SSH connections keyed by "user@host:port".
type Pool struct {
	mu      sync.Mutex
	clients map[string]*Client
}

// NewPool creates an empty connection pool.
func NewPool() *Pool {
	return &Pool{
		clients: make(map[string]*Client),
	}
}

// poolKey builds a cache key from a Config.
func poolKey(cfg Config) string {
	port := cfg.Port
	if port == 0 {
		port = DefaultPort
	}
	return fmt.Sprintf("%s@%s:%d", cfg.User, cfg.Host, port)
}

// Get returns an existing live connection for cfg, or creates a new one.
// Connections that have been closed are discarded and replaced.
func (p *Pool) Get(cfg Config) (*Client, error) {
	key := poolKey(cfg)

	p.mu.Lock()
	if c, ok := p.clients[key]; ok && !c.IsClosed() {
		p.mu.Unlock()
		return c, nil
	}
	p.mu.Unlock()

	// Create new client outside the lock to avoid holding it during dial.
	c, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	// Double-check: another goroutine may have created one while we were dialing.
	if existing, ok := p.clients[key]; ok && !existing.IsClosed() {
		p.mu.Unlock()
		// Close the one we just created; use the existing one.
		_ = c.Close()
		return existing, nil
	}
	p.clients[key] = c
	p.mu.Unlock()

	return c, nil
}

// Remove evicts a client from the pool and closes it.
func (p *Pool) Remove(cfg Config) {
	key := poolKey(cfg)
	p.mu.Lock()
	c, ok := p.clients[key]
	if ok {
		delete(p.clients, key)
	}
	p.mu.Unlock()
	if ok && c != nil {
		_ = c.Close()
	}
}

// Close closes all connections in the pool.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, c := range p.clients {
		_ = c.Close()
		delete(p.clients, key)
	}
}

// Len returns the number of connections in the pool (including potentially stale ones).
func (p *Pool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.clients)
}
