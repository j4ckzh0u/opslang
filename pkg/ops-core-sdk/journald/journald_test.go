package journald

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "journald.conf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGetDirective(t *testing.T) {
	path := setupTempConfig(t, "[Journal]\nStorage=persistent\nCompress=yes\n")
	val, err := getDirective(path, "Storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "persistent" {
		t.Errorf("got %q, want %q", val, "persistent")
	}
}

func TestGetDirectiveCaseInsensitive(t *testing.T) {
	path := setupTempConfig(t, "[Journal]\nstorage=volatile\n")
	val, err := getDirective(path, "Storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "volatile" {
		t.Errorf("got %q, want %q", val, "volatile")
	}
}

func TestGetDirectiveMissing(t *testing.T) {
	path := setupTempConfig(t, "[Journal]\n#Storage=persistent\n")
	_, err := getDirective(path, "Storage")
	if err == nil {
		t.Error("expected error for missing directive")
	}
}

func TestSetDirectiveNew(t *testing.T) {
	path := setupTempConfig(t, "[Journal]\n#Storage=persistent\n")
	if err := setDirective(path, "Storage", "persistent"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := getDirective(path, "Storage")
	if val != "persistent" {
		t.Errorf("got %q, want %q", val, "persistent")
	}
}

func TestSetDirectiveReplace(t *testing.T) {
	path := setupTempConfig(t, "[Journal]\nStorage=persistent\n")
	if err := setDirective(path, "Storage", "volatile"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := getDirective(path, "Storage")
	if val != "volatile" {
		t.Errorf("got %q, want %q", val, "volatile")
	}
}

func TestSetDirectiveNoSection(t *testing.T) {
	path := setupTempConfig(t, "")
	if err := setDirective(path, "Storage", "persistent"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "[Journal]") {
		t.Error("expected [Journal] section to be created")
	}
	val, _ := getDirective(path, "Storage")
	if val != "persistent" {
		t.Errorf("got %q, want %q", val, "persistent")
	}
}

func TestGetIgnoresOtherSections(t *testing.T) {
	path := setupTempConfig(t, "[Other]\nStorage=wrong\n[Journal]\nStorage=correct\n")
	val, err := getDirective(path, "Storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "correct" {
		t.Errorf("got %q, want %q", val, "correct")
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

func TestSetEmptyValue(t *testing.T) {
	_, err := Set("Storage", "")
	if err == nil {
		t.Error("expected error for empty value")
	}
}
