package copy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dest := filepath.Join(dir, "dest.txt")
	os.WriteFile(src, []byte("hello"), 0644)

	r := File(src, dest, "", "", "", false)
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Changed {
		t.Error("expected changed")
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "hello" {
		t.Errorf("unexpected content: %q", data)
	}

	// Second copy should not change
	r2 := File(src, dest, "", "", "", false)
	if r2.Changed {
		t.Error("should not change on identical copy")
	}
}

func TestContent(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.txt")
	r := Content("world", dest, "", "", "", false)
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Changed {
		t.Error("expected changed")
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "world" {
		t.Errorf("unexpected: %q", data)
	}

	// Same content again should not change
	r2 := Content("world", dest, "", "", "", false)
	if r2.Changed {
		t.Error("same content should not change")
	}
}

func TestFileMissing(t *testing.T) {
	r := File("/nonexistent", "/tmp/out", "", "", "", false)
	if r.Error == "" {
		t.Error("expected error for missing src")
	}
}

func TestParseMode(t *testing.T) {
	m, err := parseMode("0755")
	if err != nil {
		t.Fatal(err)
	}
	if m != 0755 {
		t.Errorf("expected 0755, got %o", m)
	}
}
