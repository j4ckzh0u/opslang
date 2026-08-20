package ssh_config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGetOption(t *testing.T) {
	path := setupTempConfig(t, "Host example\n    User testuser\n    Port 2222\n")
	val, err := getOption(path, "example", "User")
	if err != nil {
		t.Fatalf("getOption error: %v", err)
	}
	if val != "testuser" {
		t.Errorf("got %q, want %q", val, "testuser")
	}
}

func TestGetOptionMissing(t *testing.T) {
	path := setupTempConfig(t, "Host example\n    User testuser\n")
	_, err := getOption(path, "example", "Port")
	if err == nil {
		t.Error("expected error for missing option")
	}
}

func TestGetOptionMissingHost(t *testing.T) {
	path := setupTempConfig(t, "Host example\n    User testuser\n")
	_, err := getOption(path, "nonexistent", "User")
	if err == nil {
		t.Error("expected error for missing host")
	}
}

func TestSetOptionNew(t *testing.T) {
	path := setupTempConfig(t, "")
	if err := setOption(path, "example", "User", "testuser"); err != nil {
		t.Fatalf("setOption error: %v", err)
	}
	val, _ := getOption(path, "example", "User")
	if val != "testuser" {
		t.Errorf("got %q, want %q", val, "testuser")
	}
}

func TestSetOptionAdd(t *testing.T) {
	path := setupTempConfig(t, "Host example\n    User testuser\n")
	if err := setOption(path, "example", "Port", "2222"); err != nil {
		t.Fatalf("setOption error: %v", err)
	}
	val, _ := getOption(path, "example", "Port")
	if val != "2222" {
		t.Errorf("got %q, want %q", val, "2222")
	}
}

func TestSetOptionUpdate(t *testing.T) {
	path := setupTempConfig(t, "Host example\n    User testuser\n")
	if err := setOption(path, "example", "User", "newuser"); err != nil {
		t.Fatalf("setOption error: %v", err)
	}
	val, _ := getOption(path, "example", "User")
	if val != "newuser" {
		t.Errorf("got %q, want %q", val, "newuser")
	}
}

func TestRemoveOption(t *testing.T) {
	path := setupTempConfig(t, "Host example\n    User testuser\n    Port 2222\n")
	if err := removeOption(path, "example", "Port"); err != nil {
		t.Fatalf("removeOption error: %v", err)
	}
	_, err := getOption(path, "example", "Port")
	if err == nil {
		t.Error("expected error after removal")
	}
}

func TestGetEmptyHost(t *testing.T) {
	_, err := Get("", "User", "")
	if err == nil {
		t.Error("expected error for empty host")
	}
}

func TestGetEmptyOption(t *testing.T) {
	_, err := Get("example", "", "")
	if err == nil {
		t.Error("expected error for empty option")
	}
}

func TestSetEmptyHost(t *testing.T) {
	_, err := Set("", "User", "value", "")
	if err == nil {
		t.Error("expected error for empty host")
	}
}

func TestSetEmptyOption(t *testing.T) {
	_, err := Set("example", "", "value", "")
	if err == nil {
		t.Error("expected error for empty option")
	}
}

func TestSetEmptyValue(t *testing.T) {
	_, err := Set("example", "User", "", "")
	if err == nil {
		t.Error("expected error for empty value")
	}
}
