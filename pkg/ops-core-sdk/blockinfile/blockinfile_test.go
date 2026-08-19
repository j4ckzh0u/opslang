package blockinfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManageEmptyPath(t *testing.T) {
	r := Manage("", "content", "present", "", "", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty path")
	}
}

func TestManageInvalidState(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(path, []byte("hello\n"), 0644)
	r := Manage(path, "content", "invalid", "", "", "")
	if r.Status != "failed" {
		t.Error("expected failure for invalid state")
	}
}

func TestInsertAndReadBlock(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644)

	// Insert block
	r := Manage(path, "managed line 1\nmanaged line 2", "present", "", "line1", "")
	if r.Status != "success" || !r.Changed {
		t.Fatalf("insert failed: %s %s", r.Status, r.Error)
	}

	// Read block
	block, found, err := Read(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("block not found")
	}
	if !strings.Contains(block, "managed line 1") {
		t.Error("block content mismatch")
	}
}

func TestRemoveBlock(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(path, []byte("line1\n# BEGIN ANSIBLE MANAGED BLOCK\ncontent\n# END ANSIBLE MANAGED BLOCK\nline3\n"), 0644)

	r := Manage(path, "", "absent", "", "", "")
	if r.Status != "success" || !r.Changed {
		t.Fatalf("remove failed: %s %s", r.Status, r.Error)
	}

	// Verify block is gone
	_, found, _ := Read(path, "")
	if found {
		t.Error("block should have been removed")
	}
}

func TestUpdateBlock(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(path, []byte("# BEGIN ANSIBLE MANAGED BLOCK\nold content\n# END ANSIBLE MANAGED BLOCK\n"), 0644)

	r := Manage(path, "new content", "present", "", "", "")
	if r.Status != "success" || !r.Changed {
		t.Fatalf("update failed: %s %s", r.Status, r.Error)
	}

	block, found, _ := Read(path, "")
	if !found || block != "new content" {
		t.Errorf("expected 'new content', got '%s' (found=%v)", block, found)
	}
}

func TestNoChangeWhenSame(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(path, []byte("# BEGIN ANSIBLE MANAGED BLOCK\ncontent\n# END ANSIBLE MANAGED BLOCK\n"), 0644)

	r := Manage(path, "content", "present", "", "", "")
	if r.Status != "success" || r.Changed {
		t.Error("expected no change when content is the same")
	}
}

func TestRemoveNonExistentBlock(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(path, []byte("no block here\n"), 0644)

	r := Manage(path, "", "absent", "", "", "")
	if r.Status != "success" || r.Changed {
		t.Error("expected no change when removing non-existent block")
	}
}
