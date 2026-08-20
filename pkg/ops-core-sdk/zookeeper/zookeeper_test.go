package zookeeper

import (
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	// Test with invalid server (should fail gracefully)
	_, err := Connect([]string{"invalid:2181"}, 1*time.Second)
	if err == nil {
		t.Error("expected error for invalid server")
	}
}

func TestGetValidation(t *testing.T) {
	// Test with invalid server (should fail gracefully)
	r := Get("/test", []string{"invalid:2181"})
	if r.Success {
		t.Error("expected failure for invalid server")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestSetValidation(t *testing.T) {
	// Test with invalid server (should fail gracefully)
	r := Set("/test", "value", []string{"invalid:2181"})
	if r.Success {
		t.Error("expected failure for invalid server")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestCreateValidation(t *testing.T) {
	// Test with invalid server (should fail gracefully)
	r := Create("/test", "value", false, []string{"invalid:2181"})
	if r.Success {
		t.Error("expected failure for invalid server")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestDeleteValidation(t *testing.T) {
	// Test with invalid server (should fail gracefully)
	r := Delete("/test", []string{"invalid:2181"})
	if r.Success {
		t.Error("expected failure for invalid server")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestListValidation(t *testing.T) {
	// Test with invalid server (should fail gracefully)
	r := List("/", []string{"invalid:2181"})
	if r.Success {
		t.Error("expected failure for invalid server")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestExistsValidation(t *testing.T) {
	// Test with invalid server (should fail gracefully)
	r := Exists("/test", []string{"invalid:2181"})
	if r.Success {
		t.Error("expected failure for invalid server")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}
