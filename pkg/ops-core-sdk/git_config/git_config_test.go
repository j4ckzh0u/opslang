package git_config

import (
	"os"
	"os/exec"
	"testing"
)

// skipIfNoGit skips test if git is not available or not in a git repo.
func skipIfNoGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}
	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		t.Skip("not in a git repository")
	}
}

func TestGetMissingKey(t *testing.T) {
	skipIfNoGit(t)
	result, err := Get("nonexistent.key.12345", "")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if result.Value != "" {
		t.Errorf("Expected empty value for nonexistent key, got %q", result.Value)
	}
}

func TestSetAndGet(t *testing.T) {
	skipIfNoGit(t)
	// Use local scope to avoid affecting global config
	key := "test.ops.lang"
	value := "test-value"

	// Set
	setResult, err := Set(key, value, "local")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if !setResult.Changed {
		t.Error("Expected Changed=true on first set")
	}

	// Get
	getResult, err := Get(key, "local")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if getResult.Value != value {
		t.Errorf("Get() = %q, want %q", getResult.Value, value)
	}

	// Set same value again - should not change
	setResult2, err := Set(key, value, "local")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if setResult2.Changed {
		t.Error("Expected Changed=false when setting same value")
	}

	// Cleanup
	_, _ = Unset(key, "local")
}

func TestSetIdempotent(t *testing.T) {
	skipIfNoGit(t)
	key := "test.idempotent"
	value := "same-value"

	// Set twice with same value
	_, _ = Set(key, value, "local")
	result, err := Set(key, value, "local")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if result.Changed {
		t.Error("Expected Changed=false for idempotent set")
	}

	// Cleanup
	_, _ = Unset(key, "local")
}

func TestUnsetNonexistent(t *testing.T) {
	skipIfNoGit(t)
	result, err := Unset("nonexistent.key.67890", "local")
	if err != nil {
		t.Errorf("Unset() error = %v", err)
	}
	if result.Changed {
		t.Error("Expected Changed=false when unsetting nonexistent key")
	}
}

func TestList(t *testing.T) {
	skipIfNoGit(t)
	config, err := List("")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	// Should return at least some config
	if len(config) == 0 {
		t.Log("Warning: no git config entries found")
	}
}

func TestGetEmptyKey(t *testing.T) {
	_, err := Get("", "")
	if err == nil {
		t.Error("Expected error for empty key")
	}
}

func TestSetEmptyKey(t *testing.T) {
	_, err := Set("", "value", "")
	if err == nil {
		t.Error("Expected error for empty key")
	}
}

func TestUnsetEmptyKey(t *testing.T) {
	_, err := Unset("", "")
	if err == nil {
		t.Error("Expected error for empty key")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
