// Package etcd provides etcd v3 client operations.
// Supports get, set, delete, and list operations on etcd keys.
package etcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdResult represents the result of etcd operations.
type EtcdResult struct {
	Success    bool              `json:"success"`
	Key        string            `json:"key,omitempty"`
	Value      string            `json:"value,omitempty"`
	Keys       map[string]string `json:"keys,omitempty"`
	Changed    bool              `json:"changed,omitempty"`
	Revision   int64             `json:"revision,omitempty"`
	Error      string            `json:"error,omitempty"`
	Duration   int64             `json:"duration_ms"`
}

// Config holds etcd connection configuration.
type Config struct {
	Endpoints []string
	Username  string
	Password  string
	Timeout   time.Duration
}

func defaultConfig() Config {
	return Config{
		Endpoints: []string{"localhost:2379"},
		Timeout:   5 * time.Second,
	}
}

func newClient(cfg Config) (*clientv3.Client, error) {
	if len(cfg.Endpoints) == 0 {
		cfg.Endpoints = []string{"localhost:2379"}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	clientConfig := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.Timeout,
	}

	if cfg.Username != "" {
		clientConfig.Username = cfg.Username
		clientConfig.Password = cfg.Password
	}

	return clientv3.New(clientConfig)
}

// Get retrieves a value from etcd by key.
func Get(key string, endpoints []string) EtcdResult {
	start := time.Now()

	cfg := defaultConfig()
	if len(endpoints) > 0 {
		cfg.Endpoints = endpoints
	}

	cli, err := newClient(cfg)
	if err != nil {
		return EtcdResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create client: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	resp, err := cli.Get(ctx, key)
	if err != nil {
		return EtcdResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to get key: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if len(resp.Kvs) == 0 {
		return EtcdResult{
			Success:  true,
			Key:      key,
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return EtcdResult{
		Success:  true,
		Key:      key,
		Value:    string(resp.Kvs[0].Value),
		Revision: resp.Kvs[0].ModRevision,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Set sets a key-value pair in etcd.
func Set(key, value string, endpoints []string) EtcdResult {
	start := time.Now()

	cfg := defaultConfig()
	if len(endpoints) > 0 {
		cfg.Endpoints = endpoints
	}

	cli, err := newClient(cfg)
	if err != nil {
		return EtcdResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create client: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	resp, err := cli.Put(ctx, key, value)
	if err != nil {
		return EtcdResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to set key: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return EtcdResult{
		Success:  true,
		Key:      key,
		Value:    value,
		Changed:  true,
		Revision: resp.Header.Revision,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Delete removes a key from etcd.
func Delete(key string, endpoints []string) EtcdResult {
	start := time.Now()

	cfg := defaultConfig()
	if len(endpoints) > 0 {
		cfg.Endpoints = endpoints
	}

	cli, err := newClient(cfg)
	if err != nil {
		return EtcdResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create client: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	resp, err := cli.Delete(ctx, key)
	if err != nil {
		return EtcdResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to delete key: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return EtcdResult{
		Success:  true,
		Key:      key,
		Changed:  resp.Deleted > 0,
		Revision: resp.Header.Revision,
		Duration: time.Since(start).Milliseconds(),
	}
}

// List retrieves all keys with a given prefix.
func List(prefix string, endpoints []string) EtcdResult {
	start := time.Now()

	cfg := defaultConfig()
	if len(endpoints) > 0 {
		cfg.Endpoints = endpoints
	}

	cli, err := newClient(cfg)
	if err != nil {
		return EtcdResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create client: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return EtcdResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list keys: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	keys := make(map[string]string)
	for _, kv := range resp.Kvs {
		keys[string(kv.Key)] = string(kv.Value)
	}

	return EtcdResult{
		Success:  true,
		Keys:     keys,
		Duration: time.Since(start).Milliseconds(),
	}
}
