package nfs_exports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		line     string
		path     string
		hosts    []string
		options  []string
	}{
		{"/data *(rw,sync,no_subtree_check)", "/data", []string{"*"}, []string{"rw", "sync", "no_subtree_check"}},
		{"/home 192.168.1.0/24(rw)", "/home", []string{"192.168.1.0/24"}, []string{"rw"}},
		{"/tmp *", "/tmp", []string{"*"}, nil},
	}
	for _, tt := range tests {
		entry := parseLine(tt.line)
		if tt.path == "" {
			if entry != nil {
				t.Errorf("parseLine(%q) should return nil", tt.line)
			}
			continue
		}
		if entry == nil {
			t.Fatalf("parseLine(%q) returned nil", tt.line)
		}
		if entry.Path != tt.path {
			t.Errorf("parseLine(%q) path = %q, want %q", tt.line, entry.Path, tt.path)
		}
	}
}

func TestBuildLine(t *testing.T) {
	line := buildLine("/data", "*", "rw,sync")
	if line != "/data *(rw,sync)" {
		t.Errorf("buildLine = %q, want %q", line, "/data *(rw,sync)")
	}

	line2 := buildLine("/tmp", "*", "")
	if line2 != "/tmp *" {
		t.Errorf("buildLine no opts = %q, want %q", line2, "/tmp *")
	}
}

func setupTempExports(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "exports")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFindExport(t *testing.T) {
	path := setupTempExports(t, "/data *(rw,sync)\n/home 192.168.1.0/24(rw)\n")
	entry, idx := findExport(path, "/data")
	if entry == nil || idx < 0 {
		t.Fatal("expected to find /data export")
	}
	if entry.Path != "/data" {
		t.Errorf("path = %q, want %q", entry.Path, "/data")
	}
}

func TestFindExportMissing(t *testing.T) {
	path := setupTempExports(t, "/data *(rw)\n")
	_, idx := findExport(path, "/nonexistent")
	if idx >= 0 {
		t.Error("expected idx < 0 for missing export")
	}
}

func TestPresentNew(t *testing.T) {
	path := setupTempExports(t, "")
	newLine := buildLine("/data", "*", "rw,sync")
	if err := updateExport(path, "/data", newLine, -1); err != nil {
		t.Fatalf("updateExport error: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "/data") {
		t.Error("expected /data in exports")
	}
}

func TestPresentIdempotent(t *testing.T) {
	content := "/data *(rw,sync)\n"
	path := setupTempExports(t, content)
	entry, idx := findExport(path, "/data")
	if entry == nil || idx < 0 {
		t.Fatal("expected to find /data")
	}
	newLine := buildLine("/data", "*", "rw,sync")
	if entry.Line == newLine {
		// Would be idempotent - no change needed
		return
	}
	t.Error("Expected lines to match for idempotency")
}

func TestRemoveExport(t *testing.T) {
	content := "/data *(rw)\n/home *(ro)\n"
	path := setupTempExports(t, content)
	_, idx := findExport(path, "/data")
	if idx < 0 {
		t.Fatal("expected to find /data")
	}
	if err := removeExport(path, idx); err != nil {
		t.Fatalf("removeExport error: %v", err)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "/data") {
		t.Error("expected /data to be removed")
	}
	if !strings.Contains(string(data), "/home") {
		t.Error("expected /home to remain")
	}
}

func TestListExports(t *testing.T) {
	content := "/data *(rw)\n# comment\n/home 192.168.1.0/24(rw,sync)\n"
	path := setupTempExports(t, content)
	entries, err := parseExports(path)
	if err != nil {
		t.Fatalf("parseExports error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestPresentEmptyPath(t *testing.T) {
	_, err := Present("", "*", "rw")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestPresentEmptyHosts(t *testing.T) {
	_, err := Present("/data", "", "rw")
	if err == nil {
		t.Error("expected error for empty hosts")
	}
}

func TestAbsentEmptyPath(t *testing.T) {
	_, err := Absent("")
	if err == nil {
		t.Error("expected error for empty path")
	}
}
