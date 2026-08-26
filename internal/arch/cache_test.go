package arch

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// countingExecutor is a mock SSHExecutor that records how many times it
// was called and always reports the same uname output.
type countingExecutor struct {
	calls int
	out   string
	err   error
}

func (c *countingExecutor) Exec(_ context.Context, _ string) (*ExecResult, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	return &ExecResult{Stdout: c.out, ExitCode: 0}, nil
}

func newTestCache(t *testing.T, ttl time.Duration) (*Cache, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "arch-cache.json")
	return NewCache(path, ttl), path
}

func TestCacheHitAvoidsSecondProbe(t *testing.T) {
	cache, _ := newTestCache(t, DefaultTTL)
	exec := &countingExecutor{out: "x86_64"}

	id := "host1:22"
	first, err := cache.Detect(context.Background(), exec, id)
	if err != nil {
		t.Fatalf("first Detect: %v", err)
	}
	if first != "amd64" {
		t.Fatalf("first Detect = %q, want amd64", first)
	}
	second, err := cache.Detect(context.Background(), exec, id)
	if err != nil {
		t.Fatalf("second Detect: %v", err)
	}
	if second != "amd64" {
		t.Fatalf("second Detect = %q, want amd64", second)
	}
	if exec.calls != 1 {
		t.Fatalf("probe ran %d times, want 1 (second call must hit cache)", exec.calls)
	}
}

func TestCacheTTLExpiryReprobes(t *testing.T) {
	cache, _ := newTestCache(t, 0) // ttl <= 0 means never expire; use tiny positive instead
	cache.ttl = time.Nanosecond
	exec := &countingExecutor{out: "aarch64"}

	if _, err := cache.Detect(context.Background(), exec, "h:1"); err != nil {
		t.Fatalf("first Detect: %v", err)
	}
	time.Sleep(2 * time.Millisecond) // let the entry expire
	if _, err := cache.Detect(context.Background(), exec, "h:1"); err != nil {
		t.Fatalf("second Detect: %v", err)
	}
	if exec.calls != 2 {
		t.Fatalf("probe ran %d times, want 2 (entry must have expired)", exec.calls)
	}
}

func TestDetectFailureIsNotCached(t *testing.T) {
	cache, _ := newTestCache(t, DefaultTTL)
	failing := &countingExecutor{err: errors.New("connection refused")}

	if _, err := cache.Detect(context.Background(), failing, "h:1"); err == nil {
		t.Fatal("expected error from failing probe")
	}
	if failing.calls != 1 {
		t.Fatalf("failing probe calls = %d, want 1", failing.calls)
	}

	// The failure must not be cached: a healthy executor now succeeds
	// without being poisoned by the earlier error.
	ok := &countingExecutor{out: "x86_64"}
	goarch, err := cache.Detect(context.Background(), ok, "h:1")
	if err != nil || goarch != "amd64" {
		t.Fatalf("post-failure Detect = (%q, %v), want (amd64, nil)", goarch, err)
	}
	if ok.calls != 1 {
		t.Fatalf("healthy probe ran %d times, want 1", ok.calls)
	}
}

func TestPersistenceRoundTrip(t *testing.T) {
	cache, path := newTestCache(t, DefaultTTL)
	exec := &countingExecutor{out: "riscv64"}
	if _, err := cache.Detect(context.Background(), exec, "h:9"); err != nil {
		t.Fatalf("Detect: %v", err)
	}

	reloaded := NewCache(path, DefaultTTL)
	if err := reloaded.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, ok := reloaded.Get("h:9")
	if !ok || got != "riscv64" {
		t.Fatalf("reloaded cache = (%q, %v), want (riscv64, true)", got, ok)
	}
	if exec.calls != 1 {
		t.Fatalf("probe ran %d times, want 1 across instances", exec.calls)
	}
}

func TestCorruptCacheFileIsTolerated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arch-cache.json")
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}

	cache := NewCache(path, DefaultTTL)
	loadErr := cache.Load()
	if loadErr == nil {
		t.Fatal("expected corruption to be reported to the caller")
	}
	if !strings.Contains(loadErr.Error(), "corrupted") {
		t.Fatalf("load error should mention corruption, got: %v", loadErr)
	}

	// Detection still works and overwrites the corrupt file.
	exec := &countingExecutor{out: "x86_64"}
	goarch, err := cache.Detect(context.Background(), exec, "h:1")
	if err != nil || goarch != "amd64" {
		t.Fatalf("Detect after corruption = (%q, %v)", goarch, err)
	}
	if err := cache.Put("h:2", "arm64"); err != nil {
		t.Fatalf("Put must repair the file: %v", err)
	}
	fresh := NewCache(path, DefaultTTL)
	if err := fresh.Load(); err != nil {
		t.Fatalf("cache file still corrupt after Put: %v", err)
	}
}

func TestMemoryOnlyCache(t *testing.T) {
	cache := NewCache("", DefaultTTL)
	if err := cache.Load(); err != nil {
		t.Fatalf("Load on memory-only cache: %v", err)
	}
	if err := cache.Put("h:1", "amd64"); err != nil {
		t.Fatalf("Put on memory-only cache: %v", err)
	}
	if got, ok := cache.Get("h:1"); !ok || got != "amd64" {
		t.Fatalf("Get = (%q, %v)", got, ok)
	}
}
