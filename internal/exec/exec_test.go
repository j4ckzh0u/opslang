package exec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/opslang/opslang/internal/inventory"
	"github.com/opslang/opslang/internal/runner"
)

// ============================================================
// Target parsing tests
// ============================================================

func TestParseTarget(t *testing.T) {
	tests := []struct {
		input       string
		defaultUser string
		wantName    string
		wantHost    string
		wantPort    int
		wantUser    string
	}{
		{"host1", "root", "host1", "host1", 22, "root"},
		{"user@host1", "root", "user@host1", "host1", 22, "user"},
		{"user@host1:2222", "root", "user@host1:2222", "host1", 2222, "user"},
		{"host1:2222", "admin", "host1:2222", "host1", 2222, "admin"},
		{"192.168.1.1", "root", "192.168.1.1", "192.168.1.1", 22, "root"},
		{"root@10.0.0.1:22", "admin", "root@10.0.0.1:22", "10.0.0.1", 22, "root"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseTarget(tt.input, tt.defaultUser)
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", got.Host, tt.wantHost)
			}
			if got.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", got.Port, tt.wantPort)
			}
			if got.User != tt.wantUser {
				t.Errorf("User = %q, want %q", got.User, tt.wantUser)
			}
		})
	}
}

func TestParseTargets(t *testing.T) {
	hosts := []string{"host1", "user@host2", "admin@host3:2222"}
	targets := ParseTargets(hosts, "root")

	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	// First target: default user.
	if targets[0].Host != "host1" || targets[0].User != "root" || targets[0].Port != 22 {
		t.Errorf("target[0] = %+v, want host1/root/22", targets[0])
	}

	// Second target: explicit user.
	if targets[1].Host != "host2" || targets[1].User != "user" || targets[1].Port != 22 {
		t.Errorf("target[1] = %+v, want host2/user/22", targets[1])
	}

	// Third target: explicit user and port.
	if targets[2].Host != "host3" || targets[2].User != "admin" || targets[2].Port != 2222 {
		t.Errorf("target[2] = %+v, want host3/admin/2222", targets[2])
	}
}

func TestTargetsFromInventory(t *testing.T) {
	inv := &inventory.Inventory{
		Hosts: []inventory.Host{
			{Name: "web1", Host: "10.0.0.1", Port: 22, User: "deploy"},
			{Name: "db1", Host: "10.0.0.2", Port: 2222, User: "root", KeyFile: "/home/user/.ssh/id_rsa"},
		},
	}

	targets := TargetsFromInventory(inv)
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}

	if targets[0].Name != "web1" || targets[0].Host != "10.0.0.1" {
		t.Errorf("target[0] = %+v", targets[0])
	}
	if targets[1].Port != 2222 || targets[1].KeyFile != "/home/user/.ssh/id_rsa" {
		t.Errorf("target[1] = %+v", targets[1])
	}
}

// ============================================================
// Result aggregation tests
// ============================================================

func TestResultAggregation(t *testing.T) {
	tests := []struct {
		name       string
		results    map[string]*HostResult
		wantStatus string
	}{
		{
			name: "all success",
			results: map[string]*HostResult{
				"host1": {Status: "success"},
				"host2": {Status: "success"},
			},
			wantStatus: "success",
		},
		{
			name: "partial success",
			results: map[string]*HostResult{
				"host1": {Status: "success"},
				"host2": {Status: "failed"},
			},
			wantStatus: "partial",
		},
		{
			name: "all failed",
			results: map[string]*HostResult{
				"host1": {Status: "failed"},
				"host2": {Status: "failed"},
			},
			wantStatus: "failed",
		},
		{
			name:       "empty results",
			results:    map[string]*HostResult{},
			wantStatus: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allSuccess := true
			anySuccess := false
			for _, r := range tt.results {
				if r.Status == "success" {
					anySuccess = true
				} else {
					allSuccess = false
				}
			}

			var status string
			switch {
			case allSuccess && len(tt.results) > 0:
				status = "success"
			case anySuccess:
				status = "partial"
			default:
				status = "failed"
			}

			if status != tt.wantStatus {
				t.Errorf("got status %q, want %q", status, tt.wantStatus)
			}
		})
	}
}

// ============================================================
// LoadInstructions tests
// ============================================================

func TestLoadInstructions(t *testing.T) {
	// Write a test instruction file.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	pkg := runner.InstructionPackage{
		Version: "1.0",
		TaskID:  "test-123",
		Instructions: []runner.Instruction{
			{Op: "sys.cpu.usage", Assign: "cpu"},
		},
	}

	data, _ := json.MarshalIndent(pkg, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadInstructions(path)
	if err != nil {
		t.Fatalf("LoadInstructions() error = %v", err)
	}

	if loaded.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", loaded.Version)
	}
	if loaded.TaskID != "test-123" {
		t.Errorf("TaskID = %q, want test-123", loaded.TaskID)
	}
	if len(loaded.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(loaded.Instructions))
	}
	if loaded.Instructions[0].Op != "sys.cpu.usage" {
		t.Errorf("Op = %q, want sys.cpu.usage", loaded.Instructions[0].Op)
	}
}

func TestLoadInstructionsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadInstructions(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadInstructionsInvalidPackage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte(`{"version": ""}`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadInstructions(path)
	if err == nil {
		t.Error("expected error for invalid package")
	}
}

func TestLoadInstructionsMissingFile(t *testing.T) {
	_, err := LoadInstructions("/nonexistent/path/test.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ============================================================
// firstNonEmpty tests
// ============================================================

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		values []string
		want   string
	}{
		{[]string{"a", "b"}, "a"},
		{[]string{"", "b"}, "b"},
		{[]string{"", "", "c"}, "c"},
		{[]string{"", ""}, ""},
		{[]string{}, ""},
	}

	for _, tt := range tests {
		got := firstNonEmpty(tt.values...)
		if got != tt.want {
			t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.values, got, tt.want)
		}
	}
}

// ============================================================
// Runner cache path tests
// ============================================================

func TestRunnerCachePath(t *testing.T) {
	rc := newRunnerCache("/project")
	path := rc.getCachedPath("linux", "amd64")
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
	expected := filepath.Join(rc.cacheDir, "ops-runner-linux-amd64")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}

func TestRunnerCachePathArm64(t *testing.T) {
	rc := newRunnerCache("/project")
	path := rc.getCachedPath("linux", "arm64")
	expected := filepath.Join(rc.cacheDir, "ops-runner-linux-arm64")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}

func TestDefaultCacheDirFromEnv(t *testing.T) {
	t.Setenv("OPSLANG_CACHE_DIR", "/custom/cache")
	rc := newRunnerCache("/project")
	if rc.cacheDir != "/custom/cache" {
		t.Errorf("cacheDir = %q, want /custom/cache", rc.cacheDir)
	}
}

// ============================================================
// Summary JSON tests
// ============================================================

func TestSummaryJSON(t *testing.T) {
	summary := &Summary{
		TaskID:  "test-123",
		Targets: []string{"host1", "host2"},
		Status:  "success",
		Results: map[string]*HostResult{
			"host1": {
				Status: "success",
				Data:   map[string]interface{}{"key": "value"},
			},
			"host2": {
				Status: "failed",
				Error:  "connection timeout",
			},
		},
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal summary: %v", err)
	}

	var parsed Summary
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal summary: %v", err)
	}

	if parsed.TaskID != "test-123" {
		t.Errorf("TaskID = %q, want test-123", parsed.TaskID)
	}
	if parsed.Status != "success" {
		t.Errorf("Status = %q, want success", parsed.Status)
	}
	if len(parsed.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(parsed.Targets))
	}
	if parsed.Results["host1"].Status != "success" {
		t.Errorf("host1 status = %q, want success", parsed.Results["host1"].Status)
	}
	if parsed.Results["host2"].Error != "connection timeout" {
		t.Errorf("host2 error = %q, want 'connection timeout'", parsed.Results["host2"].Error)
	}
}

// ============================================================
// Executor getRunnerBinary tests (with explicit path)
// ============================================================

func TestGetRunnerBinaryExplicitPath(t *testing.T) {
	// Create a temp file to act as the runner binary.
	dir := t.TempDir()
	runnerFile := filepath.Join(dir, "my-runner")
	if err := os.WriteFile(runnerFile, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	e := &Executor{RunnerPath: runnerFile}
	path, err := e.getRunnerBinary("linux", "amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != runnerFile {
		t.Errorf("got %q, want %q", path, runnerFile)
	}
}

func TestGetRunnerBinaryExplicitPathMissing(t *testing.T) {
	e := &Executor{RunnerPath: "/nonexistent/runner"}
	_, err := e.getRunnerBinary("linux", "amd64")
	if err == nil {
		t.Error("expected error for missing runner")
	}
}
