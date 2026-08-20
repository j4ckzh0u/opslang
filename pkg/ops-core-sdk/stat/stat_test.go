package stat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStatMissing(t *testing.T) {
	r := Stat("", false, "")
	if r.Error == "" {
		t.Error("expected error for empty path")
	}
}

func TestStatNonExistent(t *testing.T) {
	r := Stat("/nonexistent/file", false, "")
	if r.Exists {
		t.Error("expected exists=false")
	}
}

func TestStatFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("hello"), 0644)

	r := Stat(f, true, "sha256")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Exists {
		t.Error("expected exists=true")
	}
	if r.IsDir {
		t.Error("expected not dir")
	}
	if r.Size != 5 {
		t.Errorf("expected size 5, got %d", r.Size)
	}
	if r.SHA256 == "" {
		t.Error("expected sha256 checksum")
	}
}

func TestStatDir(t *testing.T) {
	dir := t.TempDir()
	r := Stat(dir, false, "")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Exists {
		t.Error("expected exists")
	}
	if !r.IsDir {
		t.Error("expected isdir")
	}
}
