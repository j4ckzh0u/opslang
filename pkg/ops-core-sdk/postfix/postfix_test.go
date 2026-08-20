package postfix

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestGetFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.cf")
	content := "# comment\nmyhostname = mail.example.com\nmydomain = example.com\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	val, err := getFromFile(path, "myhostname")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "mail.example.com" {
		t.Errorf("got %q, want %q", val, "mail.example.com")
	}
}

func TestGetFromFileMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.cf")
	content := "# empty config\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := getFromFile(path, "nonexistent")
	if err == nil {
		t.Error("expected error for missing parameter")
	}
}

func TestGetFromFileComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.cf")
	content := "#myhostname = commented\nmyhostname = active\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	val, err := getFromFile(path, "myhostname")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "active" {
		t.Errorf("got %q, want %q", val, "active")
	}
}
