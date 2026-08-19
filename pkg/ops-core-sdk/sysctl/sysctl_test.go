package sysctl

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestProcSys creates a temporary directory mimicking /proc/sys structure
// and returns the cleanup function and the path to the temp directory.
func setupTestProcSys(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "sysctl-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create test directory structure
	// net/ipv4/ip_forward
	netDir := filepath.Join(tmpDir, "net", "ipv4")
	if err := os.MkdirAll(netDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create net dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(netDir, "ip_forward"), []byte("1\n"), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to write ip_forward: %v", err)
	}

	// kernel/hostname
	kernelDir := filepath.Join(tmpDir, "kernel")
	if err := os.MkdirAll(kernelDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create kernel dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(kernelDir, "hostname"), []byte("testhost"), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to write hostname: %v", err)
	}

	// vm/swappiness
	vmDir := filepath.Join(tmpDir, "vm")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create vm dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "swappiness"), []byte("60"), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to write swappiness: %v", err)
	}

	origRoot := procSysRoot
	procSysRoot = tmpDir

	return tmpDir, func() {
		procSysRoot = origRoot
		os.RemoveAll(tmpDir)
	}
}

func TestGet(t *testing.T) {
	_, cleanup := setupTestProcSys(t)
	defer cleanup()

	tests := []struct {
		name      string
		key       string
		wantValue string
		wantErr   bool
	}{
		{"ip_forward", "net.ipv4.ip_forward", "1", false},
		{"hostname", "kernel.hostname", "testhost", false},
		{"swappiness", "vm.swappiness", "60", false},
		{"nonexistent", "nonexistent.key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Value != tt.wantValue {
				t.Errorf("Get(%q) value = %q, want %q", tt.key, result.Value, tt.wantValue)
			}
			if !tt.wantErr && result.Name != tt.key {
				t.Errorf("Get(%q) name = %q, want %q", tt.key, result.Name, tt.key)
			}
		})
	}
}

func TestSet_NoChange(t *testing.T) {
	_, cleanup := setupTestProcSys(t)
	defer cleanup()

	// Set to same value - should not change
	result, err := Set("net.ipv4.ip_forward", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Changed {
		t.Error("expected Changed=false when setting same value")
	}
	if result.Name != "net.ipv4.ip_forward" {
		t.Errorf("unexpected name: %s", result.Name)
	}
	if result.Value != "1" {
		t.Errorf("unexpected value: %s", result.Value)
	}
}

func TestSet_WithChange(t *testing.T) {
	_, cleanup := setupTestProcSys(t)
	defer cleanup()

	// Change value
	result, err := Set("net.ipv4.ip_forward", "0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Changed {
		t.Error("expected Changed=true when setting different value")
	}
	if result.Value != "0" {
		t.Errorf("unexpected value: %s", result.Value)
	}

	// Verify the file was actually written
	data, err := os.ReadFile(filepath.Join(procSysRoot, "net", "ipv4", "ip_forward"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "0" {
		t.Errorf("file content = %q, want %q", string(data), "0")
	}
}

func TestSet_NonexistentKey(t *testing.T) {
	_, cleanup := setupTestProcSys(t)
	defer cleanup()

	_, err := Set("nonexistent.key", "value")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestList(t *testing.T) {
	_, cleanup := setupTestProcSys(t)
	defer cleanup()

	results, err := List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// We created 3 files
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Check that all expected keys are present
	expectedKeys := map[string]string{
		"net.ipv4.ip_forward": "1",
		"kernel.hostname":     "testhost",
		"vm.swappiness":       "60",
	}

	for _, r := range results {
		expected, ok := expectedKeys[r.Name]
		if !ok {
			t.Errorf("unexpected key: %s", r.Name)
			continue
		}
		if r.Value != expected {
			t.Errorf("key %s value = %q, want %q", r.Name, r.Value, expected)
		}
		delete(expectedKeys, r.Name)
	}

	for key := range expectedKeys {
		t.Errorf("missing key: %s", key)
	}
}

func TestNameToPath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"net.ipv4.ip_forward", "/proc/sys/net/ipv4/ip_forward"},
		{"kernel.hostname", "/proc/sys/kernel/hostname"},
		{"vm.swappiness", "/proc/sys/vm/swappiness"},
	}

	origRoot := procSysRoot
	procSysRoot = "/proc/sys"
	defer func() { procSysRoot = origRoot }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nameToPath(tt.name)
			if got != tt.want {
				t.Errorf("nameToPath(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestPathToName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/proc/sys/net/ipv4/ip_forward", "net.ipv4.ip_forward"},
		{"/proc/sys/kernel/hostname", "kernel.hostname"},
		{"/proc/sys/vm/swappiness", "vm.swappiness"},
	}

	origRoot := procSysRoot
	procSysRoot = "/proc/sys"
	defer func() { procSysRoot = origRoot }()

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := pathToName(tt.path)
			if got != tt.want {
				t.Errorf("pathToName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
