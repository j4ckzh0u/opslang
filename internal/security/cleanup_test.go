//go:build opssec

package security

import (
	"os"
	"strings"
	"testing"
)

func TestNewTempDir(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir failed: %v", err)
	}
	defer td.Cleanup()

	// Verify path is set
	path := td.Path()
	if path == "" {
		t.Error("Path() returned empty string")
	}

	// Verify path has expected prefix
	if !strings.HasPrefix(path, os.TempDir()) {
		t.Errorf("Path() = %v, expected to start with %v", path, os.TempDir())
	}
	if !strings.Contains(path, "ops-") {
		t.Errorf("Path() = %v, expected to contain 'ops-'", path)
	}

	// Verify directory exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Path is not a directory")
	}
}

func TestTempDirCleanup(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir failed: %v", err)
	}

	path := td.Path()

	// Verify directory exists before cleanup
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Directory not created: %v", err)
	}

	// Cleanup
	td.Cleanup()

	// Verify directory is removed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Directory still exists after cleanup")
	}
}

func TestTempDirCleanupIdempotent(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir failed: %v", err)
	}

	// Call cleanup multiple times - should not panic
	td.Cleanup()
	td.Cleanup()
	td.Cleanup()
}

func TestTempDirIsCleaned(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir failed: %v", err)
	}

	// Should not be cleaned initially
	if td.IsCleaned() {
		t.Error("IsCleaned() = true before Cleanup()")
	}

	// Cleanup
	td.Cleanup()

	// Should be cleaned now
	if !td.IsCleaned() {
		t.Error("IsCleaned() = false after Cleanup()")
	}
}

func TestTempDirCreateFile(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir failed: %v", err)
	}
	defer td.Cleanup()

	// Create a file in the temp directory
	filePath := td.Path() + "/test.txt"
	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create file in temp dir: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("File not created: %v", err)
	}

	// Cleanup should remove the directory and its contents
	td.Cleanup()

	// Verify directory is removed
	if _, err := os.Stat(td.Path()); !os.IsNotExist(err) {
		t.Error("Directory still exists after cleanup")
	}
}

func TestMultipleTempDirs(t *testing.T) {
	td1, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir 1 failed: %v", err)
	}
	defer td1.Cleanup()

	td2, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir 2 failed: %v", err)
	}
	defer td2.Cleanup()

	// Paths should be different
	if td1.Path() == td2.Path() {
		t.Error("Two temp dirs have the same path")
	}

	// Both should exist
	if _, err := os.Stat(td1.Path()); err != nil {
		t.Errorf("TempDir 1 not created: %v", err)
	}
	if _, err := os.Stat(td2.Path()); err != nil {
		t.Errorf("TempDir 2 not created: %v", err)
	}
}

func TestTempDirPathConsistency(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir failed: %v", err)
	}
	defer td.Cleanup()

	// Call Path() multiple times
	path1 := td.Path()
	path2 := td.Path()
	path3 := td.Path()

	// All should be the same
	if path1 != path2 || path2 != path3 {
		t.Error("Path() returns different values on different calls")
	}
}
