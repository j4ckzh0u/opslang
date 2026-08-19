package wait_for_connection

import "testing"

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
	// Very short timeout to test timeout behavior
	r := Wait("192.0.2.1", 22, 2, 1) // 2 second timeout
	if r.Status != "failed" {
		t.Error("expected failure due to timeout")
	}
	if r.Reachable {
		t.Error("should not be reachable")
	}
}
