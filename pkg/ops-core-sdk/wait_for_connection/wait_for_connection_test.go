package wait_for_connection

import (
	"net"
	"testing"
)

func TestWaitEmptyHost(t *testing.T) {
	r := Wait("", 22, 1, 1)
	if r.Status != "failed" {
		t.Error("expected failure for empty host")
	}
}

func TestCheckOnceEmptyHost(t *testing.T) {
	r := CheckOnce("", 22)
	if r.Status != "failed" {
		t.Error("expected failure for empty host")
	}
}

func TestCheckOnceInvalidHost(t *testing.T) {
	r := CheckOnce("nonexistent.invalid.host.xyz", 22)
	if r.Status != "success" {
		t.Error("expected success status even if unreachable")
	}
	if r.Reachable {
		t.Error("should not be reachable")
	}
}

func TestWaitTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Use a locally unused port so the timeout assertion is independent of routing.
	r := Wait("127.0.0.1", port, 2, 1)
	if r.Status != "failed" {
		t.Error("expected failure due to timeout")
	}
	if r.Reachable {
		t.Error("should not be reachable")
	}
}
