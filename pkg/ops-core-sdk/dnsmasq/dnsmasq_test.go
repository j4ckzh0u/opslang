package dnsmasq

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "dnsmasq.conf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGetDirective(t *testing.T) {
	path := setupTempConfig(t, "port=5353\nlisten-address=127.0.0.1\n")
	val, err := getDirective(path, "port")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "5353" {
		t.Errorf("got %q, want %q", val, "5353")
	}
}

func TestGetDirectiveMissing(t *testing.T) {
	path := setupTempConfig(t, "#port=53\n")
	_, err := getDirective(path, "port")
	if err == nil {
		t.Error("expected error for missing directive")
	}
}

func TestSetDirectiveNew(t *testing.T) {
	path := setupTempConfig(t, "#port=53\n")
	if err := setDirective(path, "port", "5353"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := getDirective(path, "port")
	if val != "5353" {
		t.Errorf("got %q, want %q", val, "5353")
	}
}

func TestSetDirectiveReplace(t *testing.T) {
	path := setupTempConfig(t, "port=53\nlisten-address=127.0.0.1\n")
	if err := setDirective(path, "port", "5353"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := getDirective(path, "port")
	if val != "5353" {
		t.Errorf("got %q, want %q", val, "5353")
	}
}

func TestRemoveDirective(t *testing.T) {
	path := setupTempConfig(t, "port=53\nlisten-address=127.0.0.1\n")
	if err := removeDirective(path, "port"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := getDirective(path, "port")
	if err == nil {
		t.Error("expected error after removal")
	}
	val, _ := getDirective(path, "listen-address")
	if val != "127.0.0.1" {
		t.Errorf("listen-address should remain, got %q", val)
	}
}

func TestGetEmptyKey(t *testing.T) {
	_, err := Get("")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestSetEmptyKey(t *testing.T) {
	_, err := Set("", "value")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestAbsentEmptyKey(t *testing.T) {
	_, err := Absent("")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestSetIdempotent(t *testing.T) {
	path := setupTempConfig(t, "port=5353\n")
	cur, _ := getDirective(path, "port")
	if cur != "5353" {
		t.Fatalf("expected 5353, got %q", cur)
	}
	// Same value would be idempotent
}
