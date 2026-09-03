package sshx

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
)

var ErrPoolClosed = errors.New("SSH connection pool is closed")

// ClientFactory creates an SSH client for a normalized configuration.
type ClientFactory func(cfg *Config) (*Client, error)

// Pool limits open SSH clients and reuses them across operations.
type Pool struct {
	mu      sync.Mutex
	max     int
	factory ClientFactory
	closed  bool
	total   int
	idle    map[poolKey][]*Client
	clients map[*Client]poolEntry
	notify  chan struct{}
}

type poolEntry struct {
	key   poolKey
	inUse bool
}

type poolKey struct {
	host         string
	port         int
	user         string
	passwordHash [sha256.Size]byte
	keyFile      string
	knownHosts   string
	insecure     bool
}

// NewPool creates a pool with a global client limit.
func NewPool(maxConnections int, factory ClientFactory) (*Pool, error) {
	if maxConnections <= 0 {
		return nil, fmt.Errorf("max connections must be positive")
	}
	if factory == nil {
		return nil, fmt.Errorf("client factory is required")
	}
	return &Pool{
		max:     maxConnections,
		factory: factory,
		idle:    make(map[poolKey][]*Client),
		clients: make(map[*Client]poolEntry),
		notify:  make(chan struct{}),
	}, nil
}

// Acquire returns a matching idle client or creates one when capacity allows.
// The caller must connect newly created clients and then call Release or Discard.
func (p *Pool) Acquire(ctx context.Context, cfg *Config) (*Client, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}
	normalized, key, err := normalizePoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	for {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return nil, ErrPoolClosed
		}
		if client := p.takeIdleLocked(key); client != nil {
			p.mu.Unlock()
			return client, nil
		}
		if p.total < p.max {
			p.total++
			p.mu.Unlock()
			return p.createClient(normalized, key)
		}
		if client := p.evictIdleLocked(); client != nil {
			p.mu.Unlock()
			if err := client.Close(); err != nil {
				return nil, fmt.Errorf("failed to close evicted SSH client: %w", err)
			}
			continue
		}
		wait := p.notify
		p.mu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-wait:
		}
	}
}

// Release returns a checked-out client to the idle pool.
func (p *Pool) Release(client *Client) error {
	if client == nil {
		return fmt.Errorf("client is required")
	}

	p.mu.Lock()
	entry, ok := p.clients[client]
	if !ok || !entry.inUse {
		p.mu.Unlock()
		return fmt.Errorf("client is not checked out from this pool")
	}
	if p.closed || !client.IsConnected() {
		delete(p.clients, client)
		p.total--
		p.signalLocked()
		p.mu.Unlock()
		return client.Close()
	}
	entry.inUse = false
	p.clients[client] = entry
	p.idle[entry.key] = append(p.idle[entry.key], client)
	p.signalLocked()
	p.mu.Unlock()
	return nil
}

// Discard removes and closes a client that cannot be reused safely.
func (p *Pool) Discard(client *Client) error {
	if client == nil {
		return fmt.Errorf("client is required")
	}
	p.mu.Lock()
	if _, ok := p.clients[client]; !ok {
		p.mu.Unlock()
		return fmt.Errorf("client does not belong to this pool")
	}
	delete(p.clients, client)
	p.removeIdleLocked(client)
	p.total--
	p.signalLocked()
	p.mu.Unlock()
	return client.Close()
}

// Close closes every idle or checked-out client and wakes blocked callers.
func (p *Pool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	clients := make([]*Client, 0, len(p.clients))
	for client := range p.clients {
		clients = append(clients, client)
	}
	p.clients = make(map[*Client]poolEntry)
	p.idle = make(map[poolKey][]*Client)
	p.total = 0
	p.signalLocked()
	p.mu.Unlock()

	var closeErrs []error
	for _, client := range clients {
		if err := client.Close(); err != nil {
			closeErrs = append(closeErrs, err)
		}
	}
	return errors.Join(closeErrs...)
}

func normalizePoolConfig(cfg *Config) (Config, poolKey, error) {
	if cfg == nil {
		return Config{}, poolKey{}, fmt.Errorf("config is required")
	}
	normalized := *cfg
	if err := normalized.Validate(); err != nil {
		return Config{}, poolKey{}, fmt.Errorf("invalid config: %w", err)
	}
	normalized.SetDefaults()
	key := poolKey{
		host:         normalized.Host,
		port:         normalized.Port,
		user:         normalized.User,
		passwordHash: sha256.Sum256([]byte(normalized.Password)),
		keyFile:      normalized.KeyFile,
		knownHosts:   normalized.KnownHostsFile,
		insecure:     normalized.InsecureSkipHostKeyVerify,
	}
	return normalized, key, nil
}

func (p *Pool) createClient(cfg Config, key poolKey) (*Client, error) {
	client, err := p.factory(&cfg)
	if err != nil || client == nil {
		p.mu.Lock()
		p.decrementTotalLocked()
		p.signalLocked()
		p.mu.Unlock()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("client factory returned nil client")
	}

	p.mu.Lock()
	if p.closed {
		p.decrementTotalLocked()
		p.signalLocked()
		p.mu.Unlock()
		return nil, errors.Join(ErrPoolClosed, client.Close())
	}
	p.clients[client] = poolEntry{key: key, inUse: true}
	p.mu.Unlock()
	return client, nil
}

func (p *Pool) takeIdleLocked(key poolKey) *Client {
	clients := p.idle[key]
	for len(clients) > 0 {
		last := len(clients) - 1
		client := clients[last]
		clients = clients[:last]
		if len(clients) == 0 {
			delete(p.idle, key)
		} else {
			p.idle[key] = clients
		}
		entry := p.clients[client]
		entry.inUse = true
		p.clients[client] = entry
		if client.IsConnected() {
			return client
		}
		delete(p.clients, client)
		p.total--
	}
	return nil
}

func (p *Pool) evictIdleLocked() *Client {
	for key, clients := range p.idle {
		if len(clients) == 0 {
			delete(p.idle, key)
			continue
		}
		last := len(clients) - 1
		client := clients[last]
		if last == 0 {
			delete(p.idle, key)
		} else {
			p.idle[key] = clients[:last]
		}
		delete(p.clients, client)
		p.total--
		return client
	}
	return nil
}

func (p *Pool) removeIdleLocked(target *Client) {
	for key, clients := range p.idle {
		for i, client := range clients {
			if client != target {
				continue
			}
			clients[i] = clients[len(clients)-1]
			clients = clients[:len(clients)-1]
			if len(clients) == 0 {
				delete(p.idle, key)
			} else {
				p.idle[key] = clients
			}
			return
		}
	}
}

func (p *Pool) signalLocked() {
	close(p.notify)
	p.notify = make(chan struct{})
}

func (p *Pool) decrementTotalLocked() {
	if p.total > 0 {
		p.total--
	}
}
