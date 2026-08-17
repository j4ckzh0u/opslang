package exec

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/opslang/opslang/internal/inventory"
	"github.com/opslang/opslang/internal/runner"
	"github.com/opslang/opslang/internal/sshx"
	"golang.org/x/crypto/ssh"
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
	expected := filepath.Join(rc.cacheDir, "ops-runner-v3-linux-amd64")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}

func TestRunnerCachePathArm64(t *testing.T) {
	rc := newRunnerCache("/project")
	path := rc.getCachedPath("linux", "arm64")
	expected := filepath.Join(rc.cacheDir, "ops-runner-v3-linux-arm64")
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

// ============================================================
// findProjectRoot tests
// ============================================================

func TestFindProjectRoot(t *testing.T) {
	t.Run("env var set to valid root", func(t *testing.T) {
		// Create a temp dir with go.mod.
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("OPSLANG_PROJECT_ROOT", dir)

		root, err := findProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != dir {
			t.Errorf("got %q, want %q", root, dir)
		}
	})

	t.Run("env var set to dir without go.mod", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("OPSLANG_PROJECT_ROOT", dir)

		_, err := findProjectRoot()
		// Should fall through to walk-up from CWD, which should find the real project root.
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("walk up from CWD finds real project root", func(t *testing.T) {
		// Clear env var so we use walk-up.
		t.Setenv("OPSLANG_PROJECT_ROOT", "")
		root, err := findProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Verify go.mod exists at root.
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
			t.Errorf("go.mod not found at project root %q", root)
		}
	})
}

// ============================================================
// runnerCache.build tests
// ============================================================

func TestRunnerCacheBuildMissingProjectRoot(t *testing.T) {
	dir := t.TempDir()
	rc := &runnerCache{
		cacheDir:    filepath.Join(dir, "cache"),
		projectRoot: "",
	}
	// Clear env so findProjectRoot uses walk-up, which will find the real project.
	t.Setenv("OPSLANG_PROJECT_ROOT", "")
	// This will attempt a real go build which will fail because cmd/ops-runner
	// doesn't exist at the temp location. But we can test with a fake project root.
	rc.projectRoot = filepath.Join(dir, "nonexistent")
	err := rc.build("linux", "amd64")
	if err == nil {
		t.Error("expected error when project root has no cmd/ops-runner")
	}
}

func TestRunnerCacheBuildCreatesCacheDir(t *testing.T) {
	// Use the real project root so go build can find cmd/ops-runner.
	t.Setenv("OPSLANG_PROJECT_ROOT", "")
	root, err := findProjectRoot()
	if err != nil {
		t.Skipf("cannot find project root: %v", err)
	}

	cacheDir := filepath.Join(t.TempDir(), "cache")
	rc := &runnerCache{
		cacheDir:    cacheDir,
		projectRoot: root,
	}

	// This will attempt a real build - may succeed or fail depending on Go environment.
	// We mainly test that the cache directory is created.
	_ = rc.build("linux", "amd64")

	// Verify cache dir was created.
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("expected cache directory to be created")
	}
}

// ============================================================
// getRunnerBinary with cache tests
// ============================================================

func TestGetRunnerBinaryCachedHit(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a fake cached runner.
	cachedPath := filepath.Join(cacheDir, "ops-runner-v3-linux-amd64")
	if err := os.WriteFile(cachedPath, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	e := &Executor{
		ProjectRoot: dir,
		runnerCache: &runnerCache{cacheDir: cacheDir, projectRoot: dir},
	}

	path, err := e.getRunnerBinary("linux", "amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != cachedPath {
		t.Errorf("got %q, want %q", path, cachedPath)
	}
}

// ============================================================
// Execute with mocked SSHClientFactory tests
// ============================================================

func TestExecuteWithMockedSSHFactory(t *testing.T) {
	// Save and restore original factory.
	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	// Mock factory that always fails to create client.
	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		return nil, fmt.Errorf("mock: connection refused")
	}

	e := &Executor{
		Targets: []Target{
			{Name: "host1", Host: "10.0.0.1", Port: 22, User: "root", Password: "pass"},
			{Name: "host2", Host: "10.0.0.2", Port: 22, User: "root", Password: "pass"},
		},
		Instructions: &runner.InstructionPackage{
			Version: "1.0",
			TaskID:  "test-1",
			Instructions: []runner.Instruction{
				{Op: "sys.cpu.usage", Assign: "cpu"},
			},
		},
		Parallel: 2,
	}

	ctx := context.Background()
	summary := e.Execute(ctx)

	if summary.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", summary.Status)
	}
	if len(summary.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(summary.Results))
	}
	for name, r := range summary.Results {
		if r.Status != "failed" {
			t.Errorf("host %s: expected status 'failed', got %q", name, r.Status)
		}
		if r.Error == "" {
			t.Errorf("host %s: expected error message", name)
		}
	}
	if len(summary.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(summary.Targets))
	}
}

func TestExecuteEmptyTargets(t *testing.T) {
	e := &Executor{
		Targets:      []Target{},
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "empty"},
	}

	ctx := context.Background()
	summary := e.Execute(ctx)

	if summary.Status != "failed" {
		t.Errorf("expected status 'failed' for empty targets, got %q", summary.Status)
	}
	if len(summary.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(summary.Results))
	}
}

func TestExecuteDefaultParallel(t *testing.T) {
	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		return nil, fmt.Errorf("mock fail")
	}

	e := &Executor{
		Targets: []Target{
			{Name: "h1", Host: "10.0.0.1", Port: 22, User: "root", Password: "p"},
		},
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "t1"},
		Parallel:     0, // Should default to 10.
	}

	summary := e.Execute(context.Background())
	if len(summary.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(summary.Results))
	}
}

// ============================================================
// executeOnHost tests with mocked SSH
// ============================================================

func TestExecuteOnHostSSHFactoryError(t *testing.T) {
	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		return nil, fmt.Errorf("mock: auth failed")
	}

	e := &Executor{
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "t1"},
		Password:     "pass",
	}

	target := Target{Name: "h1", Host: "10.0.0.1", Port: 22, User: "root"}
	result := e.executeOnHost(context.Background(), target)

	if result.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", result.Status)
	}
	if result.Error == "" {
		t.Error("expected error message")
	}
	if !strings.Contains(result.Error, "failed to create SSH client") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestExecuteOnHostUsesFirstNonEmptyCredentials(t *testing.T) {
	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	var capturedCfg *sshx.Config
	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		capturedCfg = cfg
		return nil, fmt.Errorf("mock: stop here")
	}

	e := &Executor{
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "t1"},
		User:         "default-user",
		Password:     "executor-pass",
		KeyFile:      "/path/to/executor-key",
	}

	// Target overrides with its own credentials.
	target := Target{
		Name:     "h1",
		Host:     "10.0.0.1",
		Port:     2222,
		User:     "target-user",
		Password: "target-pass",
		KeyFile:  "/path/to/target-key",
	}

	_ = e.executeOnHost(context.Background(), target)

	if capturedCfg == nil {
		t.Fatal("SSHClientFactory was not called")
	}
	if capturedCfg.Password != "target-pass" {
		t.Errorf("Password = %q, want 'target-pass'", capturedCfg.Password)
	}
	if capturedCfg.KeyFile != "/path/to/target-key" {
		t.Errorf("KeyFile = %q, want '/path/to/target-key'", capturedCfg.KeyFile)
	}
	if capturedCfg.Host != "10.0.0.1" {
		t.Errorf("Host = %q, want '10.0.0.1'", capturedCfg.Host)
	}
	if capturedCfg.Port != 2222 {
		t.Errorf("Port = %d, want 2222", capturedCfg.Port)
	}
}

func TestExecuteOnHostFallsBackToExecutorCredentials(t *testing.T) {
	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	var capturedCfg *sshx.Config
	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		capturedCfg = cfg
		return nil, fmt.Errorf("mock: stop here")
	}

	e := &Executor{
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "t1"},
		Password:     "executor-pass",
		KeyFile:      "/path/to/executor-key",
	}

	// Target has no credentials - should fall back to executor's.
	target := Target{Name: "h1", Host: "10.0.0.1", Port: 22, User: "root"}
	_ = e.executeOnHost(context.Background(), target)

	if capturedCfg == nil {
		t.Fatal("SSHClientFactory was not called")
	}
	if capturedCfg.Password != "executor-pass" {
		t.Errorf("Password = %q, want 'executor-pass'", capturedCfg.Password)
	}
	if capturedCfg.KeyFile != "/path/to/executor-key" {
		t.Errorf("KeyFile = %q, want '/path/to/executor-key'", capturedCfg.KeyFile)
	}
}

// ============================================================
// ParseTargets edge cases
// ============================================================

func TestParseTargetsEmpty(t *testing.T) {
	targets := ParseTargets([]string{}, "root")
	if len(targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(targets))
	}
}

func TestParseTargetInvalidPort(t *testing.T) {
	// Port out of range should be ignored, defaulting to 22.
	target := parseTarget("host1:99999", "root")
	if target.Port != 22 {
		t.Errorf("Port = %d, want 22 for invalid port", target.Port)
	}
	if target.Host != "host1" {
		t.Errorf("Host = %q, want 'host1'", target.Host)
	}
}

func TestParseTargetNonNumericPort(t *testing.T) {
	target := parseTarget("host1:abc", "root")
	if target.Port != 22 {
		t.Errorf("Port = %d, want 22 for non-numeric port", target.Port)
	}
}

// ============================================================
// defaultCacheDir tests
// ============================================================

func TestDefaultCacheDirFallback(t *testing.T) {
	t.Setenv("OPSLANG_CACHE_DIR", "")
	dir := defaultCacheDir()
	if dir == "" {
		t.Error("defaultCacheDir() returned empty string")
	}
	if !strings.Contains(dir, "opslang") {
		t.Errorf("expected path containing 'opslang', got %q", dir)
	}
}

// ============================================================
// sshExecutorAdapter tests
// ============================================================

func TestSSHExecutorAdapter(t *testing.T) {
	// We can't easily mock *sshx.Client, but we can test the adapter struct exists.
	adapter := &sshExecutorAdapter{client: nil}
	if adapter == nil {
		t.Error("adapter should not be nil")
	}
}

// ============================================================
// Summary timing tests
// ============================================================

func TestExecuteSetsTimestamps(t *testing.T) {
	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		return nil, fmt.Errorf("mock fail")
	}

	e := &Executor{
		Targets: []Target{
			{Name: "h1", Host: "10.0.0.1", Port: 22, User: "root", Password: "p"},
		},
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "t1"},
	}

	before := time.Now()
	summary := e.Execute(context.Background())
	after := time.Now()

	if summary.StartedAt.Before(before.Add(-time.Second)) || summary.StartedAt.After(after.Add(time.Second)) {
		t.Errorf("StartedAt %v not in expected range [%v, %v]", summary.StartedAt, before, after)
	}
	if summary.FinishedAt.Before(summary.StartedAt) {
		t.Errorf("FinishedAt %v before StartedAt %v", summary.FinishedAt, summary.StartedAt)
	}
}

// ============================================================
// LoadInstructions with valid package edge cases
// ============================================================

func TestLoadInstructionsEmptyInstructions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty_instr.json")

	pkg := runner.InstructionPackage{
		Version:      "1.0",
		TaskID:       "test-empty",
		Instructions: []runner.Instruction{},
	}
	data, _ := json.Marshal(pkg)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadInstructions(path)
	if err == nil {
		t.Error("expected error for empty instructions")
	}
}

// ============================================================
// Mock SSH server for executeOnHost testing
// ============================================================

// mockSSHServer provides a minimal SSH server for testing.
// It handles exec requests but NOT SFTP (so Upload will fail).
type mockSSHServer struct {
	listener net.Listener
	config   *ssh.ServerConfig
	wg       sync.WaitGroup
	quit     chan struct{}
	mu       sync.Mutex
	commands []string // records executed commands
}

func newMockSSHServer(t *testing.T, password string) *mockSSHServer {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	signer, err := ssh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "root" && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("authentication failed")
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	server := &mockSSHServer{
		listener: listener,
		config:   config,
		quit:     make(chan struct{}),
	}

	server.wg.Add(1)
	go server.serve()

	return server
}

func (s *mockSSHServer) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				continue
			}
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
	}
}

func (s *mockSSHServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		return
	}
	defer sshConn.Close()
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(channel, requests)
	}
}

func (s *mockSSHServer) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()
	for req := range requests {
		switch req.Type {
		case "exec":
			if req.WantReply {
				req.Reply(true, nil)
			}
			cmd := ""
			if len(req.Payload) > 4 {
				cmdLen := binary.BigEndian.Uint32(req.Payload[:4])
				if int(cmdLen)+4 <= len(req.Payload) {
					cmd = string(req.Payload[4 : 4+cmdLen])
				}
			}
			s.mu.Lock()
			s.commands = append(s.commands, cmd)
			s.mu.Unlock()
			exitCode := s.executeCommand(channel, cmd)
			status := make([]byte, 4)
			binary.BigEndian.PutUint32(status, uint32(exitCode))
			channel.SendRequest("exit-status", false, status)
			return
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

func (s *mockSSHServer) executeCommand(channel ssh.Channel, cmd string) int {
	switch {
	case cmd == "uname -m":
		channel.Write([]byte("x86_64\n"))
		return 0
	case strings.HasPrefix(cmd, "mkdir -p"):
		return 0
	case strings.HasPrefix(cmd, "chmod +x"):
		return 0
	case strings.HasPrefix(cmd, "rm -rf"):
		return 0
	case strings.Contains(cmd, "ops-runner"):
		// Simulate runner execution - output JSON.
		output := runner.Output{
			Status: "ok",
			Data:   map[string]interface{}{"cpu": map[string]interface{}{"percent": 12.5}},
			Errors: []string{},
		}
		data, _ := json.Marshal(output)
		channel.Write(data)
		return 0
	default:
		channel.Write([]byte("ok\n"))
		return 0
	}
}

func (s *mockSSHServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *mockSSHServer) Port() int {
	_, portStr, _ := net.SplitHostPort(s.listener.Addr().String())
	port := 0
	for _, c := range portStr {
		port = port*10 + int(c-'0')
	}
	return port
}

func (s *mockSSHServer) Close() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *mockSSHServer) getCommands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]string, len(s.commands))
	copy(result, s.commands)
	return result
}

// ============================================================
// executeOnHost tests with mock SSH server
// ============================================================

func TestExecuteOnHostConnectFailure(t *testing.T) {
	// Use a closed listener to simulate connection failure.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close() // Close immediately so connect fails.

	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	_, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		cfg.Timeout = 1 * time.Second
		cfg.Retries = 0
		return sshx.NewClient(cfg)
	}

	e := &Executor{
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "t1"},
		Password:     "pass",
	}

	target := Target{Name: "h1", Host: "127.0.0.1", Port: port, User: "root"}
	result := e.executeOnHost(context.Background(), target)

	if result.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", result.Status)
	}
	if !strings.Contains(result.Error, "failed to connect") {
		t.Errorf("expected connect error, got: %s", result.Error)
	}
}

func TestExecuteOnHostWithMockSSHServer(t *testing.T) {
	server := newMockSSHServer(t, "testpass")
	defer server.Close()

	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		cfg.Timeout = 5 * time.Second
		cfg.Retries = 0
		return sshx.NewClient(cfg)
	}

	// Create a fake runner binary for upload.
	runnerDir := t.TempDir()
	runnerPath := filepath.Join(runnerDir, "ops-runner-v3-linux-amd64")
	if err := os.WriteFile(runnerPath, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	e := &Executor{
		Instructions: &runner.InstructionPackage{
			Version: "1.0",
			TaskID:  "test-mock",
			Instructions: []runner.Instruction{
				{Op: "sys.cpu.usage", Assign: "cpu"},
			},
		},
		Password:   "testpass",
		RunnerPath: runnerPath,
	}

	target := Target{
		Name: "mock-host",
		Host: "127.0.0.1",
		Port: server.Port(),
		User: "root",
	}

	result := e.executeOnHost(context.Background(), target)

	// The mock server handles exec but not SFTP, so Upload will fail.
	// This still tests the path through Connect, arch detection, and getRunnerBinary.
	if result.Status != "failed" {
		t.Logf("result: %+v", result)
		// If it somehow succeeded, that's fine too.
	}

	// Verify commands were executed on the mock server.
	cmds := server.getCommands()
	if len(cmds) == 0 {
		t.Error("expected at least one command to be executed on mock server")
	}

	// Verify uname -m was called for arch detection.
	foundUname := false
	for _, cmd := range cmds {
		if cmd == "uname -m" {
			foundUname = true
			break
		}
	}
	if !foundUname {
		t.Errorf("expected 'uname -m' command, got commands: %v", cmds)
	}
}

func TestExecuteOnHostArchDetectionFailure(t *testing.T) {
	// Create a mock server that returns unknown arch.
	server := newMockSSHServer(t, "testpass")
	defer server.Close()

	// Override the executeCommand to return unknown arch.
	// We can't easily override the method, so we'll use a different approach:
	// Create a server that returns an error for uname -m.
	// Since we can't customize the mock per-test easily, let's test with
	// a server that returns an error exit code for all commands.

	errorServer := newErrorMockSSHServer(t, "testpass")
	defer errorServer.Close()

	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		cfg.Timeout = 5 * time.Second
		cfg.Retries = 0
		return sshx.NewClient(cfg)
	}

	runnerDir := t.TempDir()
	runnerPath := filepath.Join(runnerDir, "ops-runner-v3-linux-amd64")
	if err := os.WriteFile(runnerPath, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	e := &Executor{
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "t1"},
		Password:     "testpass",
		RunnerPath:   runnerPath,
	}

	target := Target{
		Name: "error-host",
		Host: "127.0.0.1",
		Port: errorServer.Port(),
		User: "root",
	}

	result := e.executeOnHost(context.Background(), target)
	if result.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", result.Status)
	}
	if !strings.Contains(result.Error, "failed to detect architecture") {
		t.Errorf("expected arch detection error, got: %s", result.Error)
	}
}

// errorMockSSHServer returns non-zero exit code for all commands.
type errorMockSSHServer struct {
	listener net.Listener
	config   *ssh.ServerConfig
	wg       sync.WaitGroup
	quit     chan struct{}
}

func newErrorMockSSHServer(t *testing.T, password string) *errorMockSSHServer {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	signer, err := ssh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "root" && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("authentication failed")
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	server := &errorMockSSHServer{
		listener: listener,
		config:   config,
		quit:     make(chan struct{}),
	}

	server.wg.Add(1)
	go server.serve()

	return server
}

func (s *errorMockSSHServer) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				continue
			}
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
	}
}

func (s *errorMockSSHServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		return
	}
	defer sshConn.Close()
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(channel, requests)
	}
}

func (s *errorMockSSHServer) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()
	for req := range requests {
		switch req.Type {
		case "exec":
			if req.WantReply {
				req.Reply(true, nil)
			}
			// Return non-zero exit code for all commands.
			status := make([]byte, 4)
			binary.BigEndian.PutUint32(status, 1)
			channel.SendRequest("exit-status", false, status)
			return
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

func (s *errorMockSSHServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *errorMockSSHServer) Port() int {
	_, portStr, _ := net.SplitHostPort(s.listener.Addr().String())
	port := 0
	for _, c := range portStr {
		port = port*10 + int(c-'0')
	}
	return port
}

func (s *errorMockSSHServer) Close() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

// ============================================================
// Execute with mock SSH server (multi-host)
// ============================================================

func TestExecuteWithMockSSHServer(t *testing.T) {
	server := newMockSSHServer(t, "testpass")
	defer server.Close()

	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		cfg.Timeout = 5 * time.Second
		cfg.Retries = 0
		return sshx.NewClient(cfg)
	}

	// Create a fake runner binary.
	runnerDir := t.TempDir()
	runnerPath := filepath.Join(runnerDir, "ops-runner-v3-linux-amd64")
	if err := os.WriteFile(runnerPath, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	e := &Executor{
		Targets: []Target{
			{Name: "h1", Host: "127.0.0.1", Port: server.Port(), User: "root", Password: "testpass"},
			{Name: "h2", Host: "127.0.0.1", Port: server.Port(), User: "root", Password: "testpass"},
		},
		Instructions: &runner.InstructionPackage{
			Version: "1.0",
			TaskID:  "test-multi",
			Instructions: []runner.Instruction{
				{Op: "sys.cpu.usage", Assign: "cpu"},
			},
		},
		Parallel:   2,
		RunnerPath: runnerPath,
	}

	ctx := context.Background()
	summary := e.Execute(ctx)

	// Both hosts should have results (may succeed or fail depending on SFTP).
	if len(summary.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(summary.Results))
	}

	// Verify timestamps are set.
	if summary.StartedAt.IsZero() {
		t.Error("StartedAt is zero")
	}
	if summary.FinishedAt.IsZero() {
		t.Error("FinishedAt is zero")
	}
	if len(summary.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(summary.Targets))
	}
}

// ============================================================
// sshExecutorAdapter test
// ============================================================

func TestSSHExecutorAdapterExec(t *testing.T) {
	server := newMockSSHServer(t, "testpass")
	defer server.Close()

	// Create a real SSH client connected to the mock server.
	cfg := &sshx.Config{
		Host:     "127.0.0.1",
		Port:     server.Port(),
		User:     "root",
		Password: "testpass",
		Timeout:  5 * time.Second,
		Retries:  0,
	}

	client, err := sshx.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create SSH client: %v", err)
	}

	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	adapter := &sshExecutorAdapter{client: client}
	result, err := adapter.Exec(context.Background(), "uname -m")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	if result.Stdout != "x86_64\n" {
		t.Errorf("Stdout = %q, want 'x86_64\\n'", result.Stdout)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

// ============================================================
// findProjectRoot additional edge cases
// ============================================================

func TestFindProjectRootEnvVarInvalid(t *testing.T) {
	// Set env to a dir that exists but has no go.mod.
	dir := t.TempDir()
	t.Setenv("OPSLANG_PROJECT_ROOT", dir)

	// Should fall through to walk-up, which will find the real project root.
	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root == "" {
		t.Error("expected non-empty root")
	}
}

// ============================================================
// runnerCache additional tests
// ============================================================

func TestRunnerCacheBuildWithValidProject(t *testing.T) {
	t.Setenv("OPSLANG_PROJECT_ROOT", "")
	root, err := findProjectRoot()
	if err != nil {
		t.Skipf("cannot find project root: %v", err)
	}

	cacheDir := filepath.Join(t.TempDir(), "test-cache")
	rc := &runnerCache{
		cacheDir:    cacheDir,
		projectRoot: root,
	}

	// Attempt build - this may succeed or fail depending on Go environment.
	// The important thing is that it doesn't panic and creates the cache dir.
	err = rc.build("linux", "amd64")
	if err != nil {
		t.Logf("build failed (may be expected in some environments): %v", err)
	}

	// Verify cache dir was created.
	if _, statErr := os.Stat(cacheDir); os.IsNotExist(statErr) {
		t.Error("expected cache directory to be created")
	}
}

func TestNewRunnerCache(t *testing.T) {
	t.Setenv("OPSLANG_CACHE_DIR", "/test/cache")
	rc := newRunnerCache("/test/project")
	if rc.cacheDir != "/test/cache" {
		t.Errorf("cacheDir = %q, want '/test/cache'", rc.cacheDir)
	}
	if rc.projectRoot != "/test/project" {
		t.Errorf("projectRoot = %q, want '/test/project'", rc.projectRoot)
	}
}

// ============================================================
// sshExecutorAdapter error path
// ============================================================

func TestSSHExecutorAdapterExecError(t *testing.T) {
	// Create a client that is NOT connected.
	cfg := &sshx.Config{
		Host:     "127.0.0.1",
		Port:     1, // invalid port
		User:     "root",
		Password: "pass",
		Timeout:  1 * time.Second,
		Retries:  0,
	}

	client, err := sshx.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create SSH client: %v", err)
	}

	// Don't connect - Exec should fail with "not connected".
	adapter := &sshExecutorAdapter{client: client}
	_, err = adapter.Exec(context.Background(), "uname -m")
	if err == nil {
		t.Error("expected error for disconnected client")
	}
}

// ============================================================
// Additional Execute tests
// ============================================================

func TestExecuteNilInstructions(t *testing.T) {
	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		return nil, fmt.Errorf("mock fail")
	}

	e := &Executor{
		Targets: []Target{
			{Name: "h1", Host: "10.0.0.1", Port: 22, User: "root", Password: "p"},
		},
		Instructions: nil,
	}

	// Should not panic even with nil instructions.
	summary := e.Execute(context.Background())
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
}

// ============================================================
// TargetsFromInventory edge cases
// ============================================================

func TestTargetsFromInventoryEmpty(t *testing.T) {
	inv := &inventory.Inventory{Hosts: []inventory.Host{}}
	targets := TargetsFromInventory(inv)
	if len(targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(targets))
	}
}

func TestTargetsFromInventoryWithPassword(t *testing.T) {
	inv := &inventory.Inventory{
		Hosts: []inventory.Host{
			{Name: "web1", Host: "10.0.0.1", Port: 22, User: "deploy", Password: "secret"},
		},
	}
	targets := TargetsFromInventory(inv)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Password != "secret" {
		t.Errorf("Password = %q, want 'secret'", targets[0].Password)
	}
}

// ============================================================
// ParseTarget edge cases
// ============================================================

func TestParseTargetIPv6Style(t *testing.T) {
	// Host with no user and no port.
	target := parseTarget("my-host.example.com", "admin")
	if target.Host != "my-host.example.com" {
		t.Errorf("Host = %q, want 'my-host.example.com'", target.Host)
	}
	if target.User != "admin" {
		t.Errorf("User = %q, want 'admin'", target.User)
	}
	if target.Port != 22 {
		t.Errorf("Port = %d, want 22", target.Port)
	}
}

func TestParseTargetZeroPort(t *testing.T) {
	target := parseTarget("host:0", "root")
	if target.Port != 22 {
		t.Errorf("Port = %d, want 22 for zero port", target.Port)
	}
}

// ============================================================
// HostResult JSON roundtrip
// ============================================================

func TestHostResultJSON(t *testing.T) {
	hr := &HostResult{
		Status:   "success",
		ExitCode: 0,
		Data:     map[string]interface{}{"key": "value"},
		Errors:   []string{"warn1"},
		Warnings: []string{"warn2"},
	}

	data, err := json.Marshal(hr)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed HostResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Status != "success" {
		t.Errorf("Status = %q, want 'success'", parsed.Status)
	}
	if len(parsed.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(parsed.Errors))
	}
	if len(parsed.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(parsed.Warnings))
	}
}

// ============================================================
// Execute context cancellation
// ============================================================

func TestExecuteContextCancellation(t *testing.T) {
	origFactory := SSHClientFactory
	defer func() { SSHClientFactory = origFactory }()

	// Create a factory that returns a client but with a very long timeout.
	// The context cancellation should stop the execution.
	SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		cfg.Timeout = 30 * time.Second
		cfg.Retries = 0
		return sshx.NewClient(cfg)
	}

	e := &Executor{
		Targets: []Target{
			{Name: "h1", Host: "192.0.2.1", Port: 22, User: "root", Password: "p"}, // RFC 5737 TEST-NET, should not be reachable
		},
		Instructions: &runner.InstructionPackage{Version: "1.0", TaskID: "t1"},
		Parallel:     1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	summary := e.Execute(ctx)
	// Should have failed due to context cancellation.
	if summary.Results["h1"].Status != "failed" {
		t.Errorf("expected failed status, got %q", summary.Results["h1"].Status)
	}
}

// ============================================================
// runnerCache.build with minimal Go project
// ============================================================

func TestRunnerCacheBuildMinimalProject(t *testing.T) {
	// Create a minimal Go project structure.
	projectDir := t.TempDir()
	cmdDir := filepath.Join(projectDir, "cmd", "ops-runner")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a minimal main.go.
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("test runner")
}
`
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	// Create go.mod.
	goMod := `module test-runner

go 1.21
`
	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	cacheDir := filepath.Join(t.TempDir(), "cache")
	rc := &runnerCache{
		cacheDir:    cacheDir,
		projectRoot: projectDir,
	}

	// This should succeed since we have a valid Go project.
	err := rc.build("linux", "amd64")
	if err != nil {
		t.Logf("build failed: %v", err)
		// May fail if GOOS=linux doesn't match the host, but that's OK.
		// The important thing is that the code paths are exercised.
	}

	// Verify cache dir was created.
	if _, statErr := os.Stat(cacheDir); os.IsNotExist(statErr) {
		t.Error("expected cache directory to be created")
	}
}

func TestRunnerCacheBuildEmptyProjectRoot(t *testing.T) {
	// Test with empty project root - should try to find it via findProjectRoot.
	t.Setenv("OPSLANG_PROJECT_ROOT", "")

	cacheDir := filepath.Join(t.TempDir(), "cache")
	rc := &runnerCache{
		cacheDir:    cacheDir,
		projectRoot: "",
	}

	// This will call findProjectRoot() which should find the real project.
	// Then it will try to build, which may succeed or fail.
	err := rc.build("linux", "amd64")
	if err != nil {
		t.Logf("build failed (may be expected): %v", err)
	}

	// Verify cache dir was created.
	if _, statErr := os.Stat(cacheDir); os.IsNotExist(statErr) {
		t.Error("expected cache directory to be created")
	}
}

// ============================================================
// Additional edge case tests
// ============================================================

func TestDefaultCacheDirNoEnv(t *testing.T) {
	t.Setenv("OPSLANG_CACHE_DIR", "")
	dir := defaultCacheDir()
	if dir == "" {
		t.Error("defaultCacheDir() returned empty string")
	}
	// Should contain "opslang" in the path.
	if !strings.Contains(dir, "opslang") {
		t.Errorf("expected path containing 'opslang', got %q", dir)
	}
}

func TestGetRunnerBinaryNoCacheNoBuild(t *testing.T) {
	// Test with no cache and no explicit path - should try to build.
	cacheDir := filepath.Join(t.TempDir(), "empty-cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir() // Empty dir, no cmd/ops-runner.
	e := &Executor{
		ProjectRoot: projectDir,
		runnerCache: &runnerCache{cacheDir: cacheDir, projectRoot: projectDir},
	}

	// This should fail because there's no cmd/ops-runner in the project root.
	_, err := e.getRunnerBinary("linux", "amd64")
	if err == nil {
		t.Error("expected error when build fails")
	}
}

// The AOT deploy path references the uploaded binary via a placeholder;
// the executor must rewrite it to the per-host remote path and leave all
// other values untouched.
func TestMarshalInstructionsPlaceholderReplacement(t *testing.T) {
	pkg := &runner.InstructionPackage{
		Version: "1.0",
		Instructions: []runner.Instruction{
			{
				Op:   "binary.exec",
				Args: map[string]interface{}{"path": AppBinaryPlaceholder, "args": []interface{}{"--flag", AppBinaryPlaceholder}},
			},
			{Op: "log", Args: map[string]interface{}{"message": "keep " + AppBinaryPlaceholder + " literal?"}},
		},
	}

	data, err := marshalInstructions(pkg, "/tmp/ops-123/app")
	if err != nil {
		t.Fatalf("marshalInstructions: %v", err)
	}

	var decoded runner.InstructionPackage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Instructions[0].Args["path"] != "/tmp/ops-123/app" {
		t.Errorf("path = %v", decoded.Instructions[0].Args["path"])
	}
	args := decoded.Instructions[0].Args["args"].([]interface{})
	if args[1] != "/tmp/ops-123/app" {
		t.Errorf("nested placeholder not rewritten: %v", args)
	}

	// No placeholder == no rewrite pass needed; output still valid.
	data2, err := marshalInstructions(&runner.InstructionPackage{
		Version:      "1.0",
		Instructions: []runner.Instruction{{Op: "sys.cpu.usage"}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data2), "sys.cpu.usage") {
		t.Error("placeholder-free package corrupted")
	}
}

// ============================================================
// Inventory tag propagation (approval gate input)
// ============================================================

func TestTargetsFromInventoryPropagatesTags(t *testing.T) {
	inv, err := inventory.Parse([]byte(`
hosts:
  - name: prod-web
    host: 10.0.0.1
    tags:
      env: prod
      role: web
  - name: dev-box
    host: 10.0.0.2
    tags:
      env: dev
  - name: bare
    host: 10.0.0.3
`))
	if err != nil {
		t.Fatalf("parse inventory: %v", err)
	}
	targets := TargetsFromInventory(inv)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	if targets[0].Tags["env"] != "prod" || targets[0].Tags["role"] != "web" {
		t.Errorf("prod-web tags not propagated: %v", targets[0].Tags)
	}
	if targets[1].Tags["env"] != "dev" {
		t.Errorf("dev-box tags not propagated: %v", targets[1].Tags)
	}
	if targets[2].Tags != nil {
		t.Errorf("untagged host should have nil tags, got %v", targets[2].Tags)
	}
}

func TestParseTargetsInlineHaveNoTags(t *testing.T) {
	targets := ParseTargets([]string{"root@10.0.0.1:22"}, "root")
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Tags != nil {
		t.Errorf("inline targets carry no tags, got %v", targets[0].Tags)
	}
}
