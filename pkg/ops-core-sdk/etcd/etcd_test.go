package etcd

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if len(cfg.Endpoints) == 0 {
		t.Error("expected default endpoints")
	}
	if cfg.Endpoints[0] != "localhost:2379" {
		t.Errorf("expected localhost:2379, got %s", cfg.Endpoints[0])
	}
	if cfg.Timeout == 0 {
		t.Error("expected default timeout")
	}
}

func TestGetValidation(t *testing.T) {
	// Test with invalid endpoint (should fail gracefully)
	r := Get("test_key", []string{"invalid:2379"})
	if r.Success {
		t.Error("expected failure for invalid endpoint")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestSetValidation(t *testing.T) {
	// Test with invalid endpoint (should fail gracefully)
	r := Set("test_key", "test_value", []string{"invalid:2379"})
	if r.Success {
		t.Error("expected failure for invalid endpoint")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestDeleteValidation(t *testing.T) {
	// Test with invalid endpoint (should fail gracefully)
	r := Delete("test_key", []string{"invalid:2379"})
	if r.Success {
		t.Error("expected failure for invalid endpoint")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestListValidation(t *testing.T) {
	// Test with invalid endpoint (should fail gracefully)
	r := List("test_prefix", []string{"invalid:2379"})
	if r.Success {
		t.Error("expected failure for invalid endpoint")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}
