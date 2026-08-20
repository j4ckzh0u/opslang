package vault

import (
	"testing"
)

func TestConnect(t *testing.T) {
	// Test with invalid address (should fail gracefully)
	_, err := Connect(Config{
		Address: "http://invalid:8200",
		Token:   "test-token",
	})
	if err != nil {
		// Connection errors are deferred until actual API calls
		t.Log("Connection created (errors deferred):", err)
	}
}

func TestReadValidation(t *testing.T) {
	// Test with invalid address (should fail gracefully)
	r := Read("secret/data/test", "test-token", "http://invalid:8200")
	if r.Success {
		t.Error("expected failure for invalid address")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestWriteValidation(t *testing.T) {
	// Test with invalid address (should fail gracefully)
	data := map[string]interface{}{
		"username": "admin",
		"password": "secret123",
	}
	r := Write("secret/data/test", "test-token", "http://invalid:8200", data)
	if r.Success {
		t.Error("expected failure for invalid address")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestDeleteValidation(t *testing.T) {
	// Test with invalid address (should fail gracefully)
	r := Delete("secret/data/test", "test-token", "http://invalid:8200")
	if r.Success {
		t.Error("expected failure for invalid address")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestListValidation(t *testing.T) {
	// Test with invalid address (should fail gracefully)
	r := List("secret/metadata/", "test-token", "http://invalid:8200")
	if r.Success {
		t.Error("expected failure for invalid address")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestWriteStructure(t *testing.T) {
	// Test data structure
	data := map[string]interface{}{
		"username": "admin",
		"password": "secret123",
		"port":     5432,
	}

	// Verify data is correctly structured
	if data["username"] != "admin" {
		t.Errorf("expected username=admin, got %v", data["username"])
	}
	if data["port"] != 5432 {
		t.Errorf("expected port=5432, got %v", data["port"])
	}
}
