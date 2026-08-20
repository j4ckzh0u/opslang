package lsb_release

import (
	"testing"
)

func TestGet(t *testing.T) {
	result, err := Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ID == "" {
		t.Error("expected non-empty ID")
	}
	if result.Kernel == "" {
		t.Error("expected non-empty Kernel")
	}
	if result.Arch == "" {
		t.Error("expected non-empty Arch")
	}
}
