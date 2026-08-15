// Package exec implements the remote execution engine for opsctl.
// It coordinates SSH connections, architecture detection, runner deployment,
// instruction execution, and result aggregation across multiple hosts.
package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/opslang/opslang/internal/arch"
	"github.com/opslang/opslang/internal/inventory"
	"github.com/opslang/opslang/internal/runner"
	"github.com/opslang/opslang/internal/sshx"
)

// Target represents a remote host to execute instructions on.
type Target struct {
	Name     string
	Host     string
	Port     int
	User     string
	Password string
	KeyFile  string
}

// HostResult captures the execution result for a single host.
type HostResult struct {
	Status     string                 `json:"status"`
	Error      string                 `json:"error,omitempty"`
	ExitCode   int                    `json:"exit_code"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Errors     []string               `json:"errors,omitempty"`
	Warnings   []string               `json:"warnings,omitempty"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
}

// Summary is the aggregated result across all hosts.
type Summary struct {
	TaskID     string                 `json:"task_id"`
	Targets    []string               `json:"targets"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
	Status     string                 `json:"status"`
	Results    map[string]*HostResult `json:"results"`
}

// Executor runs instructions on remote hosts via SSH.
type Executor struct {
	Targets      []Target
	Instructions *runner.InstructionPackage
	User         string
	KeyFile      string
	Password     string
	Parallel     int
	DryRun       bool
	RunnerPath   string // explicit path to a pre-built runner binary (optional)
	ProjectRoot  string // root of the OpsLang project (for building runner)
	runnerCache  *runnerCache
}

// SSHClientFactory creates SSH clients. Can be overridden for testing.
var SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
	return sshx.NewClient(cfg)
}

// Execute runs instructions on all targets concurrently and returns a summary.
func (e *Executor) Execute(ctx context.Context) *Summary {
	summary := &Summary{
		StartedAt: time.Now().UTC(),
		Results:   make(map[string]*HostResult),
	}

	for _, t := range e.Targets {
		summary.Targets = append(summary.Targets, t.Name)
	}

	if e.runnerCache == nil {
		e.runnerCache = newRunnerCache(e.ProjectRoot)
	}

	parallel := e.Parallel
	if parallel <= 0 {
		parallel = 10
	}
	sem := make(chan struct{}, parallel)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, target := range e.Targets {
		wg.Add(1)
		sem <- struct{}{}
		go func(t Target) {
			defer wg.Done()
			defer func() { <-sem }()

			result := e.executeOnHost(ctx, t)

			mu.Lock()
			summary.Results[t.Name] = result
			mu.Unlock()
		}(target)
	}

	wg.Wait()
	summary.FinishedAt = time.Now().UTC()

	// Determine overall status.
	allSuccess := true
	anySuccess := false
	for _, r := range summary.Results {
		if r.Status == "success" {
			anySuccess = true
		} else {
			allSuccess = false
		}
	}

	switch {
	case allSuccess && len(summary.Results) > 0:
		summary.Status = "success"
	case anySuccess:
		summary.Status = "partial"
	default:
		summary.Status = "failed"
	}

	return summary
}

// sshExecutorAdapter adapts *sshx.Client to arch.SSHExecutor.
type sshExecutorAdapter struct {
	client *sshx.Client
}

func (a *sshExecutorAdapter) Exec(ctx context.Context, cmd string) (*arch.ExecResult, error) {
	result, err := a.client.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return &arch.ExecResult{
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		ExitCode: result.ExitCode,
	}, nil
}

// executeOnHost runs the full pipeline on a single host:
// connect -> detect arch -> prepare runner -> upload -> execute -> parse output.
func (e *Executor) executeOnHost(ctx context.Context, target Target) *HostResult {
	result := &HostResult{
		StartedAt: time.Now().UTC(),
	}

	// Create SSH client.
	sshCfg := &sshx.Config{
		Host:     target.Host,
		Port:     target.Port,
		User:     target.User,
		Password: firstNonEmpty(target.Password, e.Password),
		KeyFile:  firstNonEmpty(target.KeyFile, e.KeyFile),
		Timeout:  30 * time.Second,
		Retries:  3,
	}

	client, err := SSHClientFactory(sshCfg)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to create SSH client: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Connect.
	if err := client.Connect(ctx); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to connect: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}
	defer client.Close()

	// Detect architecture.
	adapter := &sshExecutorAdapter{client: client}
	goarch, err := arch.Detect(ctx, adapter)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to detect architecture: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Find or build runner binary.
	// Remote targets are always Linux, so we build for linux/<goarch>.
	runnerPath, err := e.getRunnerBinary("linux", goarch)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to get runner binary: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Prepare remote directory.
	remoteDir := fmt.Sprintf("/tmp/ops-%d", time.Now().UnixNano())
	if _, err := client.Exec(ctx, "mkdir -p "+remoteDir); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to create remote directory: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Upload runner binary.
	remoteRunner := remoteDir + "/ops-runner"
	if err := client.Upload(ctx, runnerPath, remoteRunner); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to upload runner: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Make executable.
	if _, err := client.Exec(ctx, "chmod +x "+remoteRunner); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to chmod runner: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Marshal instruction package to JSON.
	instrJSON, err := json.Marshal(e.Instructions)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to marshal instructions: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Build runner command with optional flags.
	runnerCmd := remoteRunner
	if e.DryRun {
		runnerCmd += " --dry-run"
	}

	// Execute runner with instruction JSON piped to stdin.
	execResult, err := client.ExecWithStdin(ctx, runnerCmd, instrJSON)

	// Cleanup remote temp directory (best effort).
	_, _ = client.Exec(ctx, "rm -rf "+remoteDir)

	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to execute runner: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	if execResult.ExitCode != 0 {
		result.Status = "failed"
		result.ExitCode = execResult.ExitCode
		result.Error = fmt.Sprintf("runner exited with code %d: %s", execResult.ExitCode, execResult.Stderr)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Parse runner output.
	var output runner.Output
	if err := json.Unmarshal([]byte(execResult.Stdout), &output); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to parse runner output: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	result.ExitCode = 0
	result.Data = output.Data
	result.Errors = output.Errors
	result.Warnings = output.Warnings
	if output.Status == "ok" {
		result.Status = "success"
	} else {
		result.Status = output.Status
	}
	result.FinishedAt = time.Now().UTC()
	return result
}

// getRunnerBinary returns the path to a runner binary for the given GOOS/GOARCH.
// It checks explicit path, then cache, then builds from source.
func (e *Executor) getRunnerBinary(goos, goarch string) (string, error) {
	// If user specified an explicit runner path, use it directly.
	if e.RunnerPath != "" {
		if _, err := os.Stat(e.RunnerPath); err != nil {
			return "", fmt.Errorf("specified runner not found: %s: %w", e.RunnerPath, err)
		}
		return e.RunnerPath, nil
	}

	// Check cache.
	cached := e.runnerCache.getCachedPath(goos, goarch)
	if _, err := os.Stat(cached); err == nil {
		return cached, nil
	}

	// Build and cache.
	if err := e.runnerCache.build(goos, goarch); err != nil {
		return "", fmt.Errorf("failed to build runner: %w", err)
	}
	return cached, nil
}

// ============================================================
// Target parsing
// ============================================================

// ParseTargets parses user@host strings into Target structs.
// Accepted formats: "host", "user@host", "user@host:port".
func ParseTargets(hosts []string, defaultUser string) []Target {
	var targets []Target
	for _, h := range hosts {
		t := parseTarget(h, defaultUser)
		targets = append(targets, t)
	}
	return targets
}

// parseTarget parses a single target string.
func parseTarget(s, defaultUser string) Target {
	t := Target{
		Port: 22,
		User: defaultUser,
		Name: s, // Preserve original input as the target name.
	}

	// Extract user if present.
	if idx := strings.Index(s, "@"); idx >= 0 {
		t.User = s[:idx]
		s = s[idx+1:]
	}

	// Extract port if present.
	if idx := strings.LastIndex(s, ":"); idx >= 0 {
		portStr := s[idx+1:]
		s = s[:idx]
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil && port > 0 && port <= 65535 {
			t.Port = port
		}
	}

	t.Host = s
	return t
}

// TargetsFromInventory converts inventory hosts to targets.
func TargetsFromInventory(inv *inventory.Inventory) []Target {
	targets := make([]Target, len(inv.Hosts))
	for i, h := range inv.Hosts {
		targets[i] = Target{
			Name:     h.Name,
			Host:     h.Host,
			Port:     h.Port,
			User:     h.User,
			Password: h.Password,
			KeyFile:  h.KeyFile,
		}
	}
	return targets
}

// LoadInstructions reads and parses a JSON instruction file.
func LoadInstructions(path string) (*runner.InstructionPackage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read instruction file %s: %w", path, err)
	}

	var pkg runner.InstructionPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse instruction file: %w", err)
	}

	if err := runner.ValidatePackage(&pkg); err != nil {
		return nil, fmt.Errorf("invalid instruction package: %w", err)
	}

	return &pkg, nil
}

// ============================================================
// Runner cache
// ============================================================

type runnerCache struct {
	cacheDir    string
	projectRoot string
}

func newRunnerCache(projectRoot string) *runnerCache {
	cacheDir := defaultCacheDir()
	return &runnerCache{
		cacheDir:    cacheDir,
		projectRoot: projectRoot,
	}
}

func defaultCacheDir() string {
	if dir := os.Getenv("OPSLANG_CACHE_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	return filepath.Join(home, ".cache", "opslang", "runners")
}

// getCachedPath returns the path where a cached runner binary would be stored.
func (c *runnerCache) getCachedPath(goos, goarch string) string {
	name := fmt.Sprintf("ops-runner-%s-%s", goos, goarch)
	return filepath.Join(c.cacheDir, name)
}

// build compiles the runner from source and caches it.
func (c *runnerCache) build(goos, goarch string) error {
	root := c.projectRoot
	if root == "" {
		var err error
		root, err = findProjectRoot()
		if err != nil {
			return fmt.Errorf("cannot find project root (set OPSLANG_PROJECT_ROOT): %w", err)
		}
	}

	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	output := c.getCachedPath(goos, goarch)
	srcDir := filepath.Join(root, "cmd", "ops-runner")

	cmd := osexec.Command("go", "build", "-ldflags", "-s -w", "-o", output, ".")
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
		"CGO_ENABLED=0",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %s: %w", string(out), err)
	}

	return nil
}

// findProjectRoot locates the OpsLang project root by looking for go.mod.
func findProjectRoot() (string, error) {
	// Check environment variable first.
	if root := os.Getenv("OPSLANG_PROJECT_ROOT"); root != "" {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
	}

	// Walk up from CWD looking for go.mod.
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root (no go.mod found)")
}

// ============================================================
// Helpers
// ============================================================

// firstNonEmpty returns the first non-empty string, or empty string if all are empty.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
