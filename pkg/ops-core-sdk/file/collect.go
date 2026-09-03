// Package file - collect.go: multi-host file collection via SSH/SFTP.
//
// The Collect function orchestrates fan-in file transfers from remote hosts
// to a local destination directory. A pluggable TransferFunc performs the
// actual transfer; the default is a no-op.
package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CollectTarget describes a single remote host to collect a file from.
type CollectTarget struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	User   string `json:"user"`
	Source string `json:"source"` // remote source path
}

// CollectOptions controls the behaviour of a Collect call.
type CollectOptions struct {
	DestDir       string        `json:"dest_dir"` // local destination directory
	Parallel      int           `json:"parallel"`
	Timeout       time.Duration `json:"timeout"`
	Retries       int           `json:"retries"`
	Resume        bool          `json:"resume"`
	PartRetention time.Duration `json:"part_retention,omitempty"`
}

// CollectResult is the aggregate outcome of a Collect call.
type CollectResult struct {
	Total      int                 `json:"total"`
	Succeeded  int                 `json:"succeeded"`
	Failed     int                 `json:"failed"`
	Skipped    int                 `json:"skipped"`
	Results    []HostCollectResult `json:"results"`
	DestDir    string              `json:"dest_dir"`
	DurationMs int64               `json:"duration_ms"`
}

// HostCollectResult captures the outcome for a single source host.
type HostCollectResult struct {
	Host             string   `json:"host"`
	Status           string   `json:"status"` // success/failed/skipped
	Source           string   `json:"source"`
	Dest             string   `json:"dest"`
	Checksum         string   `json:"checksum"`
	Size             int64    `json:"size"`
	TransferSource   string   `json:"transfer_source,omitempty"`
	ResumedBytes     int64    `json:"resumed_bytes,omitempty"`
	TransferredBytes int64    `json:"transferred_bytes,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
	DurationMs       int64    `json:"duration_ms"`
	Error            string   `json:"error,omitempty"`
}

// CollectDownloadFunc is the function signature for downloading a file
// from a remote host. src is the remote path, dst is the local path.
// Implementations handle the real SSH/SFTP download; tests use mocks.
type CollectDownloadFunc func(ctx context.Context, src, dst string) error

// DefaultCollectDownloadFunc is used by Collect when no explicit download
// function is provided. The package wires a real SSH/SFTP implementation
// (see ssh_transfer.go); tests may override it.
var DefaultCollectDownloadFunc CollectDownloadFunc = func(_ context.Context, _, _ string) error {
	return fmt.Errorf("no collect download function configured; set file.DefaultCollectDownloadFunc")
}

// ResumeDownloadFunc performs a content-addressed, resumable download.
type ResumeDownloadFunc func(ctx context.Context, src, dst string, retention time.Duration) (TransferOutcome, error)

// DefaultResumeDownloadFunc is wired to the SSH/SFTP implementation by WireSSHTransfer.
var DefaultResumeDownloadFunc ResumeDownloadFunc = func(_ context.Context, _, _ string, _ time.Duration) (TransferOutcome, error) {
	return TransferOutcome{}, fmt.Errorf("no resume download function configured; set file.DefaultResumeDownloadFunc")
}

// Collect gathers files from multiple remote hosts.
// Files are organized as {destDir}/{host}/{basename}.
func Collect(source string, targets []CollectTarget, opts CollectOptions) (*CollectResult, error) {
	return CollectWith(source, targets, opts, nil)
}

// CollectWith is like Collect but accepts an explicit download function.
// If dfn is nil, DefaultCollectDownloadFunc is used.
func CollectWith(source string, targets []CollectTarget, opts CollectOptions, dfn CollectDownloadFunc) (*CollectResult, error) {
	start := time.Now()
	if opts.PartRetention < 0 {
		return nil, fmt.Errorf("file.Collect: part_retention must not be negative")
	}
	if opts.Resume && dfn != nil {
		return nil, fmt.Errorf("file.Collect: resume requires the configured resumable SSH transfer")
	}
	for index, target := range targets {
		if strings.TrimSpace(target.Host) == "" {
			return nil, fmt.Errorf("file.Collect: target %d host is empty", index)
		}
		if strings.TrimSpace(target.Source) == "" && strings.TrimSpace(source) == "" {
			return nil, fmt.Errorf("file.Collect: target %d source is empty", index)
		}
	}

	destDir := opts.DestDir
	if destDir == "" {
		destDir = filepath.Join(os.TempDir(), "ops-collect")
	}

	// Ensure destination directory exists.
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("file.Collect: create dest dir %s: %w", destDir, err)
	}

	parallel := opts.Parallel
	if parallel <= 0 {
		parallel = 5
	}
	retries := opts.Retries
	if retries <= 0 {
		retries = 3
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	downloadFn := dfn
	if downloadFn == nil {
		downloadFn = DefaultCollectDownloadFunc
	}

	result := &CollectResult{
		Total:   len(targets),
		Results: make([]HostCollectResult, 0, len(targets)),
		DestDir: destDir,
	}

	if len(targets) == 0 {
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil
	}

	sem := make(chan struct{}, parallel)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, target := range targets {
		wg.Add(1)
		go func(t CollectTarget) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			default:
				sem <- struct{}{}
			}
			defer func() { <-sem }()

			hr := HostCollectResult{
				Host:   t.Host,
				Source: t.Source,
			}

			remotePath := t.Source
			if remotePath == "" {
				remotePath = source
			}
			user := t.User
			if user == "" {
				user = "root"
			}
			port := t.Port
			if port == 0 {
				port = 22
			}
			remoteSource := formatEndpoint(user, t.Host, port, remotePath)

			// Organize as {destDir}/{host}/{basename}.
			hostDir := t.Host
			if hostDir == "" {
				hostDir = "unknown"
			}
			basename := filepath.Base(remoteSource)
			localDest := filepath.Join(destDir, hostDir, basename)

			transferStart := time.Now()
			var lastErr error
			var totalTransferred int64
			var transferWarnings []string

			for attempt := 0; attempt < retries; attempt++ {
				if attempt > 0 {
					time.Sleep(time.Duration(attempt*100) * time.Millisecond)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				var outcome TransferOutcome
				var err error
				if opts.Resume {
					outcome, err = DefaultResumeDownloadFunc(ctx, remoteSource, localDest, opts.PartRetention)
					totalTransferred += outcome.TransferredBytes
					transferWarnings = append(transferWarnings, outcome.Warnings...)
					hr.Checksum = outcome.Checksum
					hr.Size = outcome.Size
					hr.TransferSource = outcome.TransferSource
					hr.ResumedBytes = outcome.ResumedBytes
					hr.TransferredBytes = totalTransferred
					hr.Warnings = append([]string(nil), transferWarnings...)
				} else {
					err = downloadFn(ctx, remoteSource, localDest)
				}
				cancel()

				if err == nil {
					hr.Status = outcome.Status
					if hr.Status == "" {
						hr.Status = "success"
					}
					hr.Dest = localDest
					if !opts.Resume {
						hr.TransferSource = outcome.TransferSource
						hr.ResumedBytes = outcome.ResumedBytes
						hr.TransferredBytes = outcome.TransferredBytes
						hr.Warnings = append([]string(nil), outcome.Warnings...)
					}

					// Get file size and checksum of downloaded file.
					if info, statErr := os.Stat(localDest); statErr == nil {
						hr.Size = info.Size()
					}
					if outcome.Checksum != "" {
						hr.Checksum = outcome.Checksum
					} else if cs, csErr := computeFileChecksum(localDest); csErr == nil {
						hr.Checksum = cs
					}
					hr.DurationMs = time.Since(transferStart).Milliseconds()
					break
				}
				lastErr = err
			}

			if hr.Status != "success" && hr.Status != "skipped" {
				hr.Status = "failed"
				hr.DurationMs = time.Since(transferStart).Milliseconds()
				if lastErr != nil {
					hr.Error = lastErr.Error()
				}
			}

			mu.Lock()
			result.Results = append(result.Results, hr)
			switch hr.Status {
			case "success":
				result.Succeeded++
			case "failed":
				result.Failed++
			case "skipped":
				result.Skipped++
			}
			mu.Unlock()
		}(target)
	}

	wg.Wait()
	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}
