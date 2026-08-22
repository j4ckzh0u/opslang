// Package exec implements the remote execution engine for opsctl.
// It coordinates SSH connections, architecture detection, runner deployment,
// instruction execution, and result aggregation across multiple hosts.
package exec

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"sort"
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
	// Group is the inventory group the host belongs to (e.g. "root",
	// "web"). Task on-clauses may route by group name, mirroring
	// Ansible's hosts: field. Inline targets have none.
	Group string
	// Tags carries the inventory entry's tags (e.g. env=prod). Inline
	// targets have none. The approval gate classifies targets with them.
	Tags map[string]string
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

// AppBinaryPlaceholder is the instruction argument value that deploy uses to
// reference the uploaded AOT script binary. The executor replaces it with the
// real per-host remote path before sending the instruction package.
const AppBinaryPlaceholder = "@opslang-app-binary@"

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
	// AppBinary, when set, is called per host with the detected target
	// GOOS/GOARCH and must return a local binary path to upload alongside
	// the runner (AOT deploy mode). Instructions reference it via
	// AppBinaryPlaceholder.
	AppBinary func(goos, goarch string) (string, error)

	// InsecureSkipHostKeyVerify disables TOFU host key verification.
	// Lab environments only.
	InsecureSkipHostKeyVerify bool

	TaskID string

	runnerCache   *runnerCache
	buildMu       sync.Mutex // serializes runner/app binary builds
	buildInFlight map[string]*sync.Once
	buildResults  map[string]error
	appPaths      map[string]string // build key -> resolved local binary path
}

// SSHClientFactory creates SSH clients. Can be overridden for testing.
var SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
	return sshx.NewClient(cfg)
}

// Execute runs instructions on all targets concurrently and returns a summary.
func (e *Executor) Execute(ctx context.Context) *Summary {
	summary := &Summary{
		TaskID:    e.TaskID,
		StartedAt: time.Now().UTC(),
		Results:   make(map[string]*HostResult),
	}

	for _, t := range e.Targets {
		summary.Targets = append(summary.Targets, t.Name)
	}

	if e.runnerCache == nil {
		e.runnerCache = newRunnerCache(e.ProjectRoot)
	}
	if e.buildInFlight == nil {
		e.buildInFlight = make(map[string]*sync.Once)
		e.buildResults = make(map[string]error)
		e.appPaths = make(map[string]string)
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
		// Acquire the semaphore with ctx awareness so cancellation stops
		// queued hosts from starting new connections.
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			mu.Lock()
			summary.Results[target.Name] = &HostResult{
				Status:     "failed",
				Error:      fmt.Sprintf("cancelled before execution: %v", ctx.Err()),
				StartedAt:  time.Now().UTC(),
				FinishedAt: time.Now().UTC(),
			}
			mu.Unlock()
			continue
		}
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
		Host:                      target.Host,
		Port:                      target.Port,
		User:                      target.User,
		Password:                  firstNonEmpty(target.Password, e.Password),
		KeyFile:                   firstNonEmpty(target.KeyFile, e.KeyFile),
		Timeout:                   30 * time.Second,
		Retries:                   3,
		InsecureSkipHostKeyVerify: e.InsecureSkipHostKeyVerify,
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

	// AOT mode: build/fetch the compiled script binary for the detected
	// architecture and remember its local path for upload below.
	var localAppBinary string
	if e.AppBinary != nil {
		localAppBinary, err = e.getAppBinary("linux", goarch)
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("failed to build script binary for %s/%s: %v", "linux", goarch, err)
			result.FinishedAt = time.Now().UTC()
			return result
		}
	}

	// Deploy binaries through the remote content-addressed cache. The
	// runner is ~7MB; uploading it on every execution used to cost that
	// much bandwidth per host per run. Binaries are stored under
	// <cache>/<name>-<sha256[:16]> and reused when the remote checksum
	// already matches, so repeat runs transfer only a checksum query.
	remoteRunner, err := ensureRemoteBinary(ctx, client, runnerPath,
		fmt.Sprintf("ops-runner-%s-linux-%s", runnerVersionSalt, goarch))
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to stage runner: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	remoteApp := ""
	if localAppBinary != "" {
		appSum, err := fileSHA256(localAppBinary)
		if err == nil {
			remoteApp, err = ensureRemoteBinary(ctx, client, localAppBinary, "ops-app-"+appSum[:16])
		}
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("failed to stage script binary: %v", err)
			result.FinishedAt = time.Now().UTC()
			return result
		}
	}

	// Marshal instruction package to JSON, resolving the app-binary
	// placeholder to this host's remote path.
	instrJSON, err := marshalInstructions(e.Instructions, remoteApp)
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
	// Cached binaries stay on the host deliberately (they are content
	// addressed; a changed runner version gets a new path).
	execResult, err := client.ExecWithStdin(ctx, runnerCmd, instrJSON)

	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to execute runner: %v", err)
		result.FinishedAt = time.Now().UTC()
		return result
	}

	// Parse runner output. The runner exits non-zero on partial/total
	// failure but still prints its JSON report on stdout; prefer the
	// structured status when it is available.
	var output runner.Output
	parseErr := json.Unmarshal([]byte(execResult.Stdout), &output)

	if parseErr == nil && output.Status != "" {
		result.ExitCode = execResult.ExitCode
		result.Data = output.Data
		result.Errors = output.Errors
		result.Warnings = output.Warnings
		switch output.Status {
		case "ok":
			result.Status = "success"
		default:
			// partial / failed / anything else is not success.
			result.Status = output.Status
			if result.Error == "" && len(output.Errors) > 0 {
				result.Error = strings.Join(output.Errors, "; ")
			}
		}
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

	result.Status = "failed"
	result.Error = fmt.Sprintf("failed to parse runner output: %v", parseErr)
	result.FinishedAt = time.Now().UTC()
	return result
}

// marshalInstructions serializes the instruction package, replacing the
// app-binary placeholder with the host-specific remote path.
func marshalInstructions(pkg *runner.InstructionPackage, remoteApp string) ([]byte, error) {
	if pkg == nil {
		return nil, fmt.Errorf("instruction package is nil")
	}
	if remoteApp == "" || !strings.Contains(fmt.Sprint(pkg.Instructions), AppBinaryPlaceholder) {
		return json.Marshal(pkg)
	}

	// Deep-copy via JSON, then rewrite placeholder values.
	data, err := json.Marshal(pkg)
	if err != nil {
		return nil, err
	}
	var generic interface{}
	if err := json.Unmarshal(data, &generic); err != nil {
		return nil, err
	}
	replacePlaceholder(&generic, remoteApp)
	return json.Marshal(generic)
}

// replacePlaceholder recursively rewrites placeholder strings in place.
func replacePlaceholder(v *interface{}, remoteApp string) {
	switch node := (*v).(type) {
	case string:
		if node == AppBinaryPlaceholder {
			*v = remoteApp
		}
	case map[string]interface{}:
		for key, elem := range node {
			replacePlaceholder(&elem, remoteApp)
			node[key] = elem
		}
	case []interface{}:
		for i := range node {
			replacePlaceholder(&node[i], remoteApp)
		}
	}
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

	cached := e.runnerCache.getCachedPath(goos, goarch)
	if _, err := os.Stat(cached); err == nil {
		return cached, nil
	}

	if err := e.buildOnce("runner:"+goos+"/"+goarch, func() error {
		return e.runnerCache.build(goos, goarch)
	}); err != nil {
		return "", fmt.Errorf("failed to build runner: %w", err)
	}
	return cached, nil
}

// getAppBinary resolves the AOT script binary for a target architecture,
// building it at most once per (goos, goarch) across all host goroutines.
// The resolved path is shared via appPaths: storing it in a caller-local
// variable left every goroutine except the builder with an empty path.
func (e *Executor) getAppBinary(goos, goarch string) (string, error) {
	key := "app:" + goos + "/" + goarch
	err := e.buildOnce(key, func() error {
		p, err := e.AppBinary(goos, goarch)
		if err != nil {
			return err
		}
		if p == "" {
			return fmt.Errorf("app binary callback returned an empty path")
		}
		e.buildMu.Lock()
		e.appPaths[key] = p
		e.buildMu.Unlock()
		return nil
	})
	if err != nil {
		return "", err
	}
	e.buildMu.Lock()
	path := e.appPaths[key]
	e.buildMu.Unlock()
	if path == "" {
		return "", fmt.Errorf("app binary path missing for %s", key)
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("app binary %s not found: %w", path, err)
	}
	return path, nil
}

// buildOnce runs build exactly once for the given key no matter how many
// goroutines race on it; every caller observes the same result. Concurrent
// `go build` invocations writing the same output path used to corrupt
// binaries.
func (e *Executor) buildOnce(key string, build func() error) error {
	e.buildMu.Lock()
	if e.buildInFlight == nil {
		e.buildInFlight = make(map[string]*sync.Once)
	}
	if e.buildResults == nil {
		e.buildResults = make(map[string]error)
	}
	if e.appPaths == nil {
		e.appPaths = make(map[string]string)
	}
	once, exists := e.buildInFlight[key]
	if !exists {
		once = &sync.Once{}
		e.buildInFlight[key] = once
	}
	e.buildMu.Unlock()

	once.Do(func() {
		err := build()
		e.buildMu.Lock()
		e.buildResults[key] = err
		e.buildMu.Unlock()
	})

	e.buildMu.Lock()
	err := e.buildResults[key]
	e.buildMu.Unlock()
	return err
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
			Group:    h.Group,
			Tags:     h.Tags,
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

// runnerVersionSalt names the runner cache binary. It is derived from the
// content of the runner's own sources plus everything it links (the
// instruction protocol and the full op registry), so any registry change
// (new/removed operations) produces a new salt and a fresh binary — a stale
// cached runner that rejects new operation names can never be uploaded.
var runnerVersionSalt = func() string {
	h, err := hashRunnerSources()
	if err != nil {
		// Deterministic fallback: old-style manual salt. Should not
		// happen in practice; hashing only fails on unreadable trees.
		return "v3-fallback"
	}
	return fmt.Sprintf("src-%s", h[:16])
}()

// runnerSourceDirs are the trees whose content defines a runner binary.
var runnerSourceDirs = []string{
	filepath.Join("cmd", "ops-runner"),
	filepath.Join("internal", "runner"),
	filepath.Join("pkg", "ops-core-sdk"),
}

// hashRunnerSources walks the runner source trees and returns a hex sha256
// over every file path and its content, sorted for determinism.
func hashRunnerSources() (string, error) {
	root := os.Getenv("OPSLANG_PROJECT_ROOT")
	if root == "" {
		if found, err := findProjectRoot(); err == nil {
			root = found
		} else {
			wd, werr := os.Getwd()
			if werr != nil {
				return "", werr
			}
			root = wd
		}
	}

	var paths []string
	for _, dir := range runnerSourceDirs {
		err := filepath.Walk(filepath.Join(root, dir), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			paths = append(paths, path)
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	sort.Strings(paths)

	hash := sha256.New()
	for _, p := range paths {
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return "", err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		hash.Write([]byte(rel))
		hash.Write([]byte{0})
		hash.Write(data)
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getCachedPath returns the path where a cached runner binary would be stored.
func (c *runnerCache) getCachedPath(goos, goarch string) string {
	name := fmt.Sprintf("ops-runner-%s-%s-%s", runnerVersionSalt, goos, goarch)
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
// Remote binary cache
// ============================================================

// remoteCacheDir is the per-host cache for staged binaries. Content
// addressing (sha256 in the file name) makes it safe to share between
// runs; OPSLANG_REMOTE_CACHE_DIR overrides it (used by tests).
func remoteCacheDir() string {
	if d := os.Getenv("OPSLANG_REMOTE_CACHE_DIR"); d != "" {
		return d
	}
	return "/tmp/.opslang-cache"
}

// ensureRemoteBinary makes localPath available on the host at a
// content-addressed cache path and returns that path. When the remote
// file already exists with a matching checksum, no file bytes are
// transferred: the checksum is computed ON the remote host (sha256sum),
// so a cache-hit probe costs one exec channel round trip (~100 bytes),
// not a download of the whole binary.
func ensureRemoteBinary(ctx context.Context, client *sshx.Client, localPath, name string) (string, error) {
	localSum, err := fileSHA256(localPath)
	if err != nil {
		return "", fmt.Errorf("hash %s: %w", localPath, err)
	}
	remotePath := remoteCacheDir() + "/" + name + "-" + localSum[:16]

	// Cache hit probe: remote-side hashing. (Reading the file via SFTP to
	// hash it locally would transfer the full binary - the opposite of
	// what the cache is for.)
	if remoteSum, ok := remoteSHA256(ctx, client, remotePath); ok {
		if strings.EqualFold(remoteSum, localSum) {
			return remotePath, nil
		}
	}

	if _, err := client.Exec(ctx, "mkdir -p "+remoteCacheDir()); err != nil {
		return "", fmt.Errorf("create remote cache dir: %w", err)
	}
	if err := client.Upload(ctx, localPath, remotePath); err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}
	if _, err := client.Exec(ctx, "chmod 0755 "+remotePath); err != nil {
		return "", fmt.Errorf("chmod: %w", err)
	}

	// Verify what landed: a truncated upload must not silently execute.
	if remoteSum, ok := remoteSHA256(ctx, client, remotePath); !ok || !strings.EqualFold(remoteSum, localSum) {
		return "", fmt.Errorf("post-upload checksum mismatch for %s", remotePath)
	}
	return remotePath, nil
}

// remoteSHA256 hashes a file on the remote host via sha256sum. Returns
// ok=false when the utility or the file is unavailable - callers then
// (re)upload, which is always safe.
func remoteSHA256(ctx context.Context, client *sshx.Client, path string) (string, bool) {
	res, err := client.Exec(ctx, "sha256sum "+path)
	if err != nil || res.ExitCode != 0 {
		return "", false
	}
	fields := strings.Fields(res.Stdout)
	if len(fields) < 1 || len(fields[0]) != 64 {
		return "", false
	}
	return fields[0], true
}

// fileSHA256 hashes a local file.
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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
