package find

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindNoPaths(t *testing.T) {
	r := Find(FindOptions{})
	if r.Error == "" {
		t.Error("expected error for empty paths")
	}
}

func TestFindFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("2"), 0644)
	os.WriteFile(filepath.Join(dir, "c.go"), []byte("3"), 0644)

	r := Find(FindOptions{Paths: []string{dir}, Patterns: []string{"*.txt"}, FileType: "file"})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Matched != 2 {
		t.Errorf("expected 2 matched, got %d", r.Matched)
	}
}

func TestFindRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(dir, "root.txt"), []byte("r"), 0644)
	os.WriteFile(filepath.Join(sub, "nested.txt"), []byte("n"), 0644)

	r := Find(FindOptions{Paths: []string{dir}, Recurse: true})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Matched < 2 {
		t.Errorf("expected at least 2 (dir + 2 files), got %d", r.Matched)
	}

	r2 := Find(FindOptions{Paths: []string{dir}, Recurse: false})
	if r2.Error != "" {
		t.Fatal(r2.Error)
	}
	// non-recursive should miss nested.txt
	if r2.Matched >= r.Matched {
		t.Errorf("non-recursive should have fewer matches: %d vs %d", r2.Matched, r.Matched)
	}
}

func TestFindDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)
	r := Find(FindOptions{Paths: []string{dir}, FileType: "directory"})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Matched < 1 {
		t.Error("expected at least 1 directory")
	}
}
