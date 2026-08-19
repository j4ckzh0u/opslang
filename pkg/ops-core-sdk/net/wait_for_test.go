package opsnet

import (
	"fmt"
	goNet "net"
	"testing"
)

func TestWaitFor_EmptyHost(t *testing.T) {
	_, err := WaitFor("", 80, 5)
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestWaitFor_InvalidPort(t *testing.T) {
	_, err := WaitFor("localhost", 0, 5)
	if err == nil {
		t.Fatal("expected error for port 0")
	}
	_, err = WaitFor("localhost", 70000, 5)
	if err == nil {
		t.Fatal("expected error for port 70000")
	}
}

func TestWaitFor_InvalidTimeout(t *testing.T) {
	_, err := WaitFor("localhost", 80, 0)
	if err == nil {
		t.Fatal("expected error for timeout 0")
	}
}

func TestWaitFor_SuccessOnOpenPort(t *testing.T) {
	// Start a TCP listener on a random port
	ln, err := goNet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*goNet.TCPAddr).Port
	res, err := WaitFor("127.0.0.1", port, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success=true, got error: %s", res.Error)
	}
	if res.DurationMs < 0 {
		t.Fatalf("expected positive duration, got %d", res.DurationMs)
	}
}

func TestWaitFor_TimeoutOnClosedPort(t *testing.T) {
	// Find a port that's unlikely to be in use
	ln, err := goNet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	port := ln.Addr().(*goNet.TCPAddr).Port
	ln.Close() // Close immediately so nothing is listening

	res, err := WaitFor("127.0.0.1", port, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Success {
		t.Fatal("expected success=false on closed port")
	}
	if res.Error == "" {
		t.Fatal("expected error message in result")
	}
	// Duration should be at least 2 seconds (the timeout)
	if res.DurationMs < 1500 {
		t.Fatalf("expected duration >= 1500ms, got %d", res.DurationMs)
	}
	_ = fmt.Sprintf("duration was %dms", res.DurationMs)
}
