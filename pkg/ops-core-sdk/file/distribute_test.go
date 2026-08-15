package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestDistribute_Success(t *testing.T) {
	// Create a source file.
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	content := "hello distribute"
	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Mock transfer function that tracks calls.
	var callCount int64
	mockTransfer := func(_ context.Context, s, d string) error {
		atomic.AddInt64(&callCount, 1)
		if s != src {
			return fmt.Errorf("unexpected source: %s", s)
		}
		return nil
	}

	targets := []DistributeTarget{
		{Host: "host1", Port: 22, User: "root", Dest: "/tmp/dest/"},
		{Host: "host2", Port: 22, User: "root", Dest: "/tmp/dest/"},
		{Host: "host3", Port: 22, User: "root", Dest: "/tmp/dest/"},
	}

	result, err := DistributeWith(src, targets, DistributeOptions{
		Parallel: 2,
		Retries:  1,
		Checksum: true,
	}, mockTransfer)
	if err != nil {
		t.Fatalf("DistributeWith: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Total = %d, want 3", result.Total)
	}
	if result.Succeeded != 3 {
		t.Errorf("Succeeded = %d, want 3", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("Failed = %d, want 0", result.Failed)
	}
	if len(result.Results) != 3 {
		t.Errorf("len(Results) = %d, want 3", len(result.Results))
	}
	if atomic.LoadInt64(&callCount) != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
	for _, hr := range result.Results {
		if hr.Status != "success" {
			t.Errorf("host %s status = %q, want success", hr.Host, hr.Status)
		}
		if !hr.Changed {
			t.Errorf("host %s Changed = false, want true", hr.Host)
		}
		if hr.Checksum == "" {
			t.Errorf("host %s Checksum is empty (checksum=true)", hr.Host)
		}
		if hr.Size != int64(len(content)) {
			t.Errorf("host %s Size = %d, want %d", hr.Host, hr.Size, len(content))
		}
	}
	if result.DurationMs < 0 {
		t.Errorf("DurationMs = %d, want >= 0", result.DurationMs)
	}
}

func TestDistribute_Failed(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	mockTransfer := func(_ context.Context, _, _ string) error {
		return fmt.Errorf("connection refused")
	}

	targets := []DistributeTarget{
		{Host: "host1", Dest: "/tmp/"},
	}

	result, err := DistributeWith(src, targets, DistributeOptions{
		Retries: 1,
	}, mockTransfer)
	if err != nil {
		t.Fatalf("DistributeWith: %v", err)
	}

	if result.Succeeded != 0 {
		t.Errorf("Succeeded = %d, want 0", result.Succeeded)
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1", result.Failed)
	}
	if result.Results[0].Error == "" {
		t.Error("expected error message in result")
	}
}

func TestDistribute_EmptyTargets(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := Distribute(src, nil, DistributeOptions{})
	if err != nil {
		t.Fatalf("Distribute: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("Total = %d, want 0", result.Total)
	}
}

func TestDistribute_InvalidSource(t *testing.T) {
	_, err := Distribute("/nonexistent/file.txt", []DistributeTarget{
		{Host: "host1", Dest: "/tmp/"},
	}, DistributeOptions{})
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
}

func TestDistribute_Retries(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	var attempts int64
	mockTransfer := func(_ context.Context, _, _ string) error {
		n := atomic.AddInt64(&attempts, 1)
		if n <= 2 {
			return fmt.Errorf("transient error")
		}
		return nil // succeed on 3rd try
	}

	targets := []DistributeTarget{
		{Host: "host1", Dest: "/tmp/"},
	}

	result, err := DistributeWith(src, targets, DistributeOptions{
		Retries: 3,
	}, mockTransfer)
	if err != nil {
		t.Fatalf("DistributeWith: %v", err)
	}

	if result.Succeeded != 1 {
		t.Errorf("Succeeded = %d, want 1", result.Succeeded)
	}
	if atomic.LoadInt64(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestDistribute_Timeout(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	mockTransfer := func(ctx context.Context, _, _ string) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	}

	targets := []DistributeTarget{
		{Host: "host1", Dest: "/tmp/"},
	}

	start := time.Now()
	result, err := DistributeWith(src, targets, DistributeOptions{
		Timeout: 100 * time.Millisecond,
		Retries: 0,
	}, mockTransfer)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("DistributeWith: %v", err)
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1", result.Failed)
	}
	if elapsed > 2*time.Second {
		t.Errorf("elapsed = %v, expected timeout ~100ms", elapsed)
	}
}

func TestDistribute_DefaultTransferFunc(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// DefaultTransferFunc returns an error by default, so Distribute should
	// succeed at the orchestration level but each target should fail.
	result, err := Distribute(src, []DistributeTarget{
		{Host: "host1", Dest: "/tmp/"},
	}, DistributeOptions{Retries: 0})
	if err != nil {
		t.Fatalf("Distribute: %v", err)
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1 (default transfer func should fail)", result.Failed)
	}
}

func TestIsDir(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/tmp/dest/", true},
		{"/tmp/dest", true},
		{"/tmp/file.txt", false},
		{"/tmp/archive.tar.gz", false},
		{"relative/path/", true},
		{"relative/file.go", false},
	}
	for _, tt := range tests {
		got := isDir(tt.path)
		if got != tt.want {
			t.Errorf("isDir(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestComputeFileChecksum(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(p, []byte("hello"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cs1, err := computeFileChecksum(p)
	if err != nil {
		t.Fatalf("computeFileChecksum: %v", err)
	}
	if cs1 == "" {
		t.Fatal("checksum is empty")
	}

	// Same content should produce same checksum.
	cs2, err := computeFileChecksum(p)
	if err != nil {
		t.Fatalf("computeFileChecksum 2: %v", err)
	}
	if cs1 != cs2 {
		t.Errorf("checksums differ for same file: %q vs %q", cs1, cs2)
	}
}
