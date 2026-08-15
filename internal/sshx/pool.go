package sshx

import (
	"context"
	"fmt"
	"sync"
)

// Pool manages a pool of SSH clients for concurrent operations.
type Pool struct {
	config    *Config
	maxSize   int
	mu        sync.Mutex
	available []*Client
	all       []*Client
}

// NewPool creates a connection pool with the given configuration.
func NewPool(cfg *Config, maxSize int) (*Pool, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("pool size must be positive")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cfg.SetDefaults()

	return &Pool{
		config:    cfg,
		maxSize:   maxSize,
		available: make([]*Client, 0, maxSize),
		all:       make([]*Client, 0, maxSize),
	}, nil
}

// Get retrieves an available client from the pool or creates a new one.
func (p *Pool) Get(ctx context.Context) (*Client, error) {
	p.mu.Lock()
	if len(p.available) > 0 {
		client := p.available[len(p.available)-1]
		p.available = p.available[:len(p.available)-1]
		p.mu.Unlock()
		return client, nil
	}

	if len(p.all) >= p.maxSize {
		p.mu.Unlock()
		return nil, fmt.Errorf("connection pool exhausted (max: %d)", p.maxSize)
	}
	p.mu.Unlock()

	client, err := NewClient(p.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	p.mu.Lock()
	p.all = append(p.all, client)
	p.mu.Unlock()

	return client, nil
}

// Put returns a client to the pool for reuse.
func (p *Pool) Put(client *Client) {
	if client == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.available = append(p.available, client)
}

// Close closes all connections in the pool.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for _, client := range p.all {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.all = nil
	p.available = nil
	return firstErr
}

// Size returns the total number of connections in the pool.
func (p *Pool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.all)
}

// Available returns the number of available (idle) connections.
func (p *Pool) Available() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available)
}
