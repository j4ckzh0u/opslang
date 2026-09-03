package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCollect_Success(t *testing.T) {
	destDir := t.TempDir()

	// Mock download function that creates the local file.
	var callCount int64
	mockDownload := func(_ context.Context, src, dst string) error {
		atomic.AddInt64(&callCount, 1)
		// Create the local file to simulate a successful download.
		dir := filepath.Dir(dst)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return os.WriteFile(dst, []byte("collected data"), 0644)
	}

	targets := []CollectTarget{
		{Host: "host1", Port: 22, User: "root", Source: "/var/log/app.log"},
		{Host: "host2", Port: 22, User: "root", Source: "/var/log/app.log"},
	}

	result, err := CollectWith("/var/log/app.log", targets, CollectOptions{
		DestDir:  destDir,
		Parallel: 2,
		Retries:  1,
	}, mockDownload)
	if err != nil {
		t.Fatalf("CollectWith: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
	if result.Succeeded != 2 {
		t.Errorf("Succeeded = %d, want 2", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("Failed = %d, want 0", result.Failed)
	}
	if result.DestDir != destDir {
		t.Errorf("DestDir = %q, want %q", result.DestDir, destDir)
	}
	if len(result.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(result.Results))
	}

	for _, hr := range result.Results {
		if hr.Status != "success" {
			t.Errorf("host %s status = %q, want success", hr.Host, hr.Status)
		}
		if hr.Dest == "" {
			t.Errorf("host %s Dest is empty", hr.Host)
		}
		// Verify file exists on disk.
		if _, err := os.Stat(hr.Dest); err != nil {
			t.Errorf("host %s dest file missing: %v", hr.Host, err)
		}
		if hr.Size != 14 { // len("collected data")
			t.Errorf("host %s Size = %d, want 14", hr.Host, hr.Size)
		}
		if hr.Checksum == "" {
			t.Errorf("host %s Checksum is empty", hr.Host)
		}
	}

	// Verify directory structure: {destDir}/{host}/{basename}.
	for _, host := range []string{"host1", "host2"} {
		expected := filepath.Join(destDir, host, "app.log")
		if _, err := os.Stat(expected); err != nil {
			t.Errorf("expected file %s missing: %v", expected, err)
		}
	}
}

func TestCollect_Failed(t *testing.T) {
	destDir := t.TempDir()

	mockDownload := func(_ context.Context, _, _ string) error {
		return fmt.Errorf("connection refused")
	}

	targets := []CollectTarget{
		{Host: "host1", Source: "/var/log/app.log"},
	}

	result, err := CollectWith("/var/log/app.log", targets, CollectOptions{
		DestDir: destDir,
		Retries: 1,
	}, mockDownload)
	if err != nil {
		t.Fatalf("CollectWith: %v", err)
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

func TestCollect_EmptyTargets(t *testing.T) {
	destDir := t.TempDir()

	result, err := Collect("/var/log/app.log", nil, CollectOptions{
		DestDir: destDir,
	})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("Total = %d, want 0", result.Total)
	}
}

func TestCollect_DefaultDestDir(t *testing.T) {
	// When DestDir is empty, should use a default temp path.
	mockDownload := func(_ context.Context, src, dst string) error {
		dir := filepath.Dir(dst)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return os.WriteFile(dst, []byte("data"), 0644)
	}

	targets := []CollectTarget{
		{Host: "host1", Source: "/var/log/app.log"},
	}

	result, err := CollectWith("/var/log/app.log", targets, CollectOptions{
		Retries: 0,
	}, mockDownload)
	if err != nil {
		t.Fatalf("CollectWith: %v", err)
	}
	if result.DestDir == "" {
		t.Error("DestDir should not be empty")
	}
	if result.Succeeded != 1 {
		t.Errorf("Succeeded = %d, want 1", result.Succeeded)
	}
}

func TestCollect_Retries(t *testing.T) {
	destDir := t.TempDir()

	var attempts int64
	mockDownload := func(_ context.Context, src, dst string) error {
		n := atomic.AddInt64(&attempts, 1)
		if n <= 2 {
			return fmt.Errorf("transient error")
		}
		dir := filepath.Dir(dst)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return os.WriteFile(dst, []byte("data"), 0644)
	}

	targets := []CollectTarget{
		{Host: "host1", Source: "/var/log/app.log"},
	}

	result, err := CollectWith("/var/log/app.log", targets, CollectOptions{
		DestDir: destDir,
		Retries: 3,
	}, mockDownload)
	if err != nil {
		t.Fatalf("CollectWith: %v", err)
	}

	if result.Succeeded != 1 {
		t.Errorf("Succeeded = %d, want 1", result.Succeeded)
	}
	if atomic.LoadInt64(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestCollect_Timeout(t *testing.T) {
	destDir := t.TempDir()

	mockDownload := func(ctx context.Context, _, _ string) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	}

	targets := []CollectTarget{
		{Host: "host1", Source: "/var/log/app.log"},
	}

	start := time.Now()
	result, err := CollectWith("/var/log/app.log", targets, CollectOptions{
		DestDir: destDir,
		Timeout: 100 * time.Millisecond,
		Retries: 0,
	}, mockDownload)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("CollectWith: %v", err)
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1", result.Failed)
	}
	if elapsed > 2*time.Second {
		t.Errorf("elapsed = %v, expected timeout ~100ms", elapsed)
	}
}

func TestCollect_TargetOverridesSource(t *testing.T) {
	destDir := t.TempDir()

	var receivedSource string
	mockDownload := func(_ context.Context, src, dst string) error {
		receivedSource = src
		dir := filepath.Dir(dst)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return os.WriteFile(dst, []byte("data"), 0644)
	}

	targets := []CollectTarget{
		{Host: "host1", Source: "/custom/path.log"},
	}

	_, err := CollectWith("/default/path.log", targets, CollectOptions{
		DestDir: destDir,
		Retries: 0,
	}, mockDownload)
	if err != nil {
		t.Fatalf("CollectWith: %v", err)
	}

	if receivedSource != "ssh://root@host1:22/custom/path.log" {
		t.Errorf("source = %q, want ssh://root@host1:22/custom/path.log", receivedSource)
	}
}

func TestCollect_ResumableOutcome(t *testing.T) {
	destDir := t.TempDir()
	saved := DefaultResumeDownloadFunc
	DefaultResumeDownloadFunc = func(_ context.Context, endpoint, dst string, retention time.Duration) (TransferOutcome, error) {
		if !strings.Contains(endpoint, "host1") || !strings.HasPrefix(dst, destDir) {
			return TransferOutcome{}, fmt.Errorf("unexpected resumable download arguments")
		}
		if retention != 2*time.Hour {
			return TransferOutcome{}, fmt.Errorf("retention = %v", retention)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return TransferOutcome{}, err
		}
		if err := os.WriteFile(dst, []byte("data"), 0600); err != nil {
			return TransferOutcome{}, err
		}
		return TransferOutcome{
			Status:           "success",
			Changed:          true,
			Checksum:         strings.Repeat("b", 64),
			Size:             4,
			TransferSource:   "controller_sftp",
			ResumedBytes:     2,
			TransferredBytes: 2,
		}, nil
	}
	defer func() { DefaultResumeDownloadFunc = saved }()

	result, err := Collect("/var/log/app.log", []CollectTarget{{Host: "host1"}}, CollectOptions{
		DestDir:       destDir,
		Resume:        true,
		Retries:       1,
		PartRetention: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if result.Succeeded != 1 || result.Failed != 0 || result.Skipped != 0 {
		t.Fatalf("aggregate result = %+v", result)
	}
	host := result.Results[0]
	if host.ResumedBytes != 2 || host.TransferredBytes != 2 || host.TransferSource != "controller_sftp" || host.Checksum == "" {
		t.Fatalf("host result = %+v", host)
	}
}

func TestCollect_RejectsInvalidResumeConfiguration(t *testing.T) {
	if _, err := Collect("/tmp/source", nil, CollectOptions{PartRetention: -time.Second}); err == nil || !strings.Contains(err.Error(), "part_retention") {
		t.Fatalf("negative retention error = %v", err)
	}
	_, err := CollectWith("/tmp/source", nil, CollectOptions{Resume: true}, func(context.Context, string, string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "resume") {
		t.Fatalf("custom resume download error = %v", err)
	}
}
