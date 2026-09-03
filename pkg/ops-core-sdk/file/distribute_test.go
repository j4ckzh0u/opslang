package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	// Mock transfer function that tracks calls and captures endpoints.
	var callCount int64
	var mu sync.Mutex
	endpoints := map[string]bool{}
	mockTransfer := func(_ context.Context, s, d string) error {
		atomic.AddInt64(&callCount, 1)
		if s != src {
			return fmt.Errorf("unexpected source: %s", s)
		}
		mu.Lock()
		endpoints[d] = true
		mu.Unlock()
		return nil
	}

	// Stub the verify hook: Checksum:true must trigger remote verification.
	var verified int64
	savedVerify := DefaultVerifyFunc
	DefaultVerifyFunc = func(_ context.Context, endpoint, want string) error {
		if !strings.HasPrefix(endpoint, "ssh://root@") {
			return fmt.Errorf("unexpected endpoint %q", endpoint)
		}
		if want == "" {
			return fmt.Errorf("verify called without expected digest")
		}
		atomic.AddInt64(&verified, 1)
		return nil
	}
	defer func() { DefaultVerifyFunc = savedVerify }()

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
	if atomic.LoadInt64(&verified) != 3 {
		t.Errorf("verified = %d, want 3 (checksum verification must run per host)", atomic.LoadInt64(&verified))
	}
	mu.Lock()
	ep := "ssh://root@host1:22/tmp/dest/" + filepath.Base(src)
	mu.Unlock()
	if !endpoints[ep] {
		t.Errorf("expected endpoint %q among %v", ep, endpoints)
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

func TestDistribute_ResumableOutcome(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("resume data"), 0600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	saved := DefaultResumeUploadFunc
	DefaultResumeUploadFunc = func(_ context.Context, gotSource, endpoint string, retention time.Duration) (TransferOutcome, error) {
		if gotSource != src || !strings.Contains(endpoint, "host1") {
			return TransferOutcome{}, fmt.Errorf("unexpected resumable upload arguments")
		}
		if retention != time.Hour {
			return TransferOutcome{}, fmt.Errorf("retention = %v", retention)
		}
		return TransferOutcome{
			Status:           "skipped",
			Checksum:         strings.Repeat("a", 64),
			Size:             11,
			TransferSource:   "controller_sftp",
			ResumedBytes:     7,
			TransferredBytes: 4,
			Warnings:         []string{"reused partial file"},
		}, nil
	}
	defer func() { DefaultResumeUploadFunc = saved }()

	result, err := Distribute(src, []DistributeTarget{{Host: "host1", Dest: "/tmp/file"}}, DistributeOptions{
		Resume:        true,
		Retries:       1,
		PartRetention: time.Hour,
	})
	if err != nil {
		t.Fatalf("Distribute: %v", err)
	}
	if result.Skipped != 1 || result.Succeeded != 0 || result.Failed != 0 {
		t.Fatalf("aggregate result = %+v", result)
	}
	host := result.Results[0]
	if host.Changed || host.ResumedBytes != 7 || host.TransferredBytes != 4 || host.TransferSource != "controller_sftp" {
		t.Fatalf("host result = %+v", host)
	}
	if len(host.Warnings) != 1 || host.Checksum == "" {
		t.Fatalf("host observability fields = %+v", host)
	}
}

func TestDistribute_RejectsInvalidResumeConfiguration(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("data"), 0600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if _, err := Distribute(src, nil, DistributeOptions{PartRetention: -time.Second}); err == nil || !strings.Contains(err.Error(), "part_retention") {
		t.Fatalf("negative retention error = %v", err)
	}
	_, err := DistributeWith(src, nil, DistributeOptions{Resume: true}, func(context.Context, string, string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "resume") {
		t.Fatalf("custom resume transfer error = %v", err)
	}
}
