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
	Compress bool          `json:"compress"`
	DestDir  string        `json:"dest_dir"` // local destination directory
	Parallel int           `json:"parallel"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
}

// CollectResult is the aggregate outcome of a Collect call.
type CollectResult struct {
	Total      int                 `json:"total"`
	Succeeded  int                 `json:"succeeded"`
	Failed     int                 `json:"failed"`
	Results    []HostCollectResult `json:"results"`
	DestDir    string              `json:"dest_dir"`
	DurationMs int64               `json:"duration_ms"`
}

// HostCollectResult captures the outcome for a single source host.
type HostCollectResult struct {
	Host       string `json:"host"`
	Status     string `json:"status"` // success/failed/skipped
	Source     string `json:"source"`
	Dest       string `json:"dest"`
	Checksum   string `json:"checksum"`
	Size       int64  `json:"size"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
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

// Collect gathers files from multiple remote hosts.
// Files are organized as {destDir}/{host}/{basename}.
func Collect(source string, targets []CollectTarget, opts CollectOptions) (*CollectResult, error) {
	return CollectWith(source, targets, opts, nil)
}

// CollectWith is like Collect but accepts an explicit download function.
// If dfn is nil, DefaultCollectDownloadFunc is used.
func CollectWith(source string, targets []CollectTarget, opts CollectOptions, dfn CollectDownloadFunc) (*CollectResult, error) {
	start := time.Now()

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

			for attempt := 0; attempt < retries; attempt++ {
				if attempt > 0 {
					time.Sleep(time.Duration(attempt*100) * time.Millisecond)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				err := downloadFn(ctx, remoteSource, localDest)
				cancel()

				if err == nil {
					hr.Status = "success"
					hr.Dest = localDest

					// Get file size and checksum of downloaded file.
					if info, statErr := os.Stat(localDest); statErr == nil {
						hr.Size = info.Size()
					}
					if cs, csErr := computeFileChecksum(localDest); csErr == nil {
						hr.Checksum = cs
					}
					hr.DurationMs = time.Since(transferStart).Milliseconds()
					break
				}
				lastErr = err
			}

			if hr.Status != "success" {
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
			}
			mu.Unlock()
		}(target)
	}

	wg.Wait()
	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}
