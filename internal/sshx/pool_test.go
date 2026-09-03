package sshx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolReusesConnectedClient(t *testing.T) {
	server, err := newMockSSHServer("testpass")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	host, _, err := net.SplitHostPort(server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	var created atomic.Int32
	pool, err := NewPool(1, func(cfg *Config) (*Client, error) {
		created.Add(1)
		return NewClient(cfg)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	cfg := &Config{
		Host:                      host,
		Port:                      server.Port(),
		User:                      "root",
		Password:                  "testpass",
		InsecureSkipHostKeyVerify: true,
	}
	first, err := pool.Acquire(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Connect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := pool.Release(first); err != nil {
		t.Fatal(err)
	}

	second, err := pool.Acquire(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if second != first {
		t.Fatal("Acquire() did not reuse the connected client")
	}
	result, err := second.Exec(context.Background(), "echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if result.Stdout != "hello\n" {
		t.Fatalf("stdout = %q, want %q", result.Stdout, "hello\n")
	}
	if err := pool.Release(second); err != nil {
		t.Fatal(err)
	}
	if got := created.Load(); got != 1 {
		t.Fatalf("factory calls = %d, want 1", got)
	}
}

func TestPoolHonorsCapacityAndContext(t *testing.T) {
	pool, err := NewPool(1, NewClient)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &Config{Host: "example.test", Password: "secret", InsecureSkipHostKeyVerify: true}
	client, err := pool.Acquire(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := pool.Acquire(ctx, cfg); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Acquire() error = %v, want deadline exceeded", err)
	}
	if err := pool.Discard(client); err != nil {
		t.Fatal(err)
	}
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPoolSeparatesAuthenticationConfigurations(t *testing.T) {
	server, err := newMockSSHServer("testpass")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	host, _, err := net.SplitHostPort(server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	var created atomic.Int32
	pool, err := NewPool(1, func(cfg *Config) (*Client, error) {
		created.Add(1)
		return NewClient(cfg)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	firstCfg := &Config{Host: host, Port: server.Port(), User: "root", Password: "testpass", InsecureSkipHostKeyVerify: true}
	first, err := pool.Acquire(context.Background(), firstCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Connect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := pool.Release(first); err != nil {
		t.Fatal(err)
	}

	secondCfg := *firstCfg
	secondCfg.Password = "different-password"
	second, err := pool.Acquire(context.Background(), &secondCfg)
	if err != nil {
		t.Fatal(err)
	}
	if second == first {
		t.Fatal("Acquire() reused a client across different credentials")
	}
	if got := created.Load(); got != 2 {
		t.Fatalf("factory calls = %d, want 2", got)
	}
	if err := pool.Discard(second); err != nil {
		t.Fatal(err)
	}
}

func TestPoolCloseWakesWaiter(t *testing.T) {
	pool, err := NewPool(1, NewClient)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &Config{Host: "example.test", Password: "secret", InsecureSkipHostKeyVerify: true}
	client, err := pool.Acquire(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := pool.Acquire(context.Background(), cfg)
		done <- err
	}()
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, ErrPoolClosed) {
			t.Fatalf("Acquire() error = %v, want ErrPoolClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked Acquire() was not woken by Close()")
	}
	if client.IsConnected() {
		t.Fatal("Close() left a client connected")
	}
}

func TestPoolRejectsInvalidInputs(t *testing.T) {
	if _, err := NewPool(0, NewClient); err == nil {
		t.Fatal("NewPool() accepted zero capacity")
	}
	if _, err := NewPool(1, nil); err == nil {
		t.Fatal("NewPool() accepted nil factory")
	}
	pool, err := NewPool(1, NewClient)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()
	if _, err := pool.Acquire(context.Background(), nil); err == nil {
		t.Fatal("Acquire() accepted nil config")
	}
	if _, err := pool.Acquire(nil, &Config{Host: "example.test", Password: "secret"}); err == nil {
		t.Fatal("Acquire() accepted nil context")
	}
}

func TestPoolReleasesCapacityAfterFactoryError(t *testing.T) {
	var calls atomic.Int32
	pool, err := NewPool(1, func(cfg *Config) (*Client, error) {
		if calls.Add(1) == 1 {
			return nil, fmt.Errorf("temporary factory failure")
		}
		return NewClient(cfg)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()
	cfg := &Config{Host: "example.test", Password: "secret", InsecureSkipHostKeyVerify: true}
	if _, err := pool.Acquire(context.Background(), cfg); err == nil {
		t.Fatal("Acquire() expected factory error")
	}
	client, err := pool.Acquire(context.Background(), cfg)
	if err != nil {
		t.Fatalf("second Acquire() error = %v", err)
	}
	if err := pool.Discard(client); err != nil {
		t.Fatal(err)
	}
}
