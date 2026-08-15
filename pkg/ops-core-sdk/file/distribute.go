// Package file - distribute.go: multi-host file distribution via SSH/SFTP.
//
// The Distribute function orchestrates fan-out file transfers to remote hosts.
// A pluggable TransferFunc performs the actual transfer; the default is a no-op.
// Applications should set DefaultTransferFunc to a real SSH/SFTP implementation
// before calling Distribute.
package file

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DistributeTarget describes a single remote host to receive a file.
type DistributeTarget struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Dest string `json:"dest"` // remote destination path
}

// DistributeOptions controls the behaviour of a Distribute call.
type DistributeOptions struct {
	Compress bool          `json:"compress"`
	Checksum bool          `json:"checksum"`    // verify after transfer
	Mode     string        `json:"mode"`        // file permissions e.g. "0644"
	Owner    string        `json:"owner"`
	Parallel int           `json:"parallel"`    // max concurrent transfers
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
}

// DistributeResult is the aggregate outcome of a Distribute call.
type DistributeResult struct {
	Total      int                    `json:"total"`
	Succeeded  int                    `json:"succeeded"`
	Failed     int                    `json:"failed"`
	Skipped    int                    `json:"skipped"`
	Results    []HostDistributeResult `json:"results"`
	DurationMs int64                  `json:"duration_ms"`
}

// HostDistributeResult captures the outcome for a single target host.
type HostDistributeResult struct {
	Host       string `json:"host"`
	Status     string `json:"status"` // success/failed/skipped
	Changed    bool   `json:"changed"`
	Checksum   string `json:"checksum"`
	Size       int64  `json:"size"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// TransferFunc is the function signature for actual file transfer operations.
// src is the local path, dst is the remote path. Implementations handle the
// real SSH/SFTP transfer; tests use mocks.
type TransferFunc func(ctx context.Context, src, dst string) error

// DefaultTransferFunc is used by Distribute and Collect when no explicit
// transfer function is provided. Applications should set this at startup.
var DefaultTransferFunc TransferFunc = func(_ context.Context, _, _ string) error {
	return fmt.Errorf("no transfer function configured; set file.DefaultTransferFunc")
}

// Distribute sends a file to multiple remote hosts.
// If opts.Parallel <= 0, defaults to 5. If opts.Retries <= 0, defaults to 3.
// If opts.Timeout <= 0, defaults to 60s per transfer.
func Distribute(source string, targets []DistributeTarget, opts DistributeOptions) (*DistributeResult, error) {
	return DistributeWith(source, targets, opts, nil)
}

// DistributeWith is like Distribute but accepts an explicit transfer function.
// If tfn is nil, DefaultTransferFunc is used.
func DistributeWith(source string, targets []DistributeTarget, opts DistributeOptions, tfn TransferFunc) (*DistributeResult, error) {
	start := time.Now()

	// Validate source exists.
	srcInfo, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("file.Distribute: source %s: %w", source, err)
	}
	srcSize := srcInfo.Size()

	// Apply defaults.
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

	transferFn := tfn
	if transferFn == nil {
		transferFn = DefaultTransferFunc
	}

	result := &DistributeResult{
		Total:   len(targets),
		Results: make([]HostDistributeResult, 0, len(targets)),
	}

	if len(targets) == 0 {
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil
	}

	// Pre-compute source checksum if requested.
	var srcChecksum string
	if opts.Checksum {
		srcChecksum, err = computeFileChecksum(source)
		if err != nil {
			return nil, fmt.Errorf("file.Distribute: compute source checksum: %w", err)
		}
	}

	sem := make(chan struct{}, parallel)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, target := range targets {
		wg.Add(1)
		go func(t DistributeTarget) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			default:
				// Semaphore full, wait.
				sem <- struct{}{}
			}
			defer func() { <-sem }()

			hr := HostDistributeResult{
				Host: t.Host,
			}
			transferStart := time.Now()

			// Determine remote path.
			remotePath := t.Dest
			if remotePath == "" {
				remotePath = filepath.Join("/tmp", filepath.Base(source))
			} else if isDir(remotePath) {
				remotePath = filepath.Join(remotePath, filepath.Base(source))
			}

			var lastErr error
			for attempt := 0; attempt <= retries; attempt++ {
				if attempt > 0 {
					time.Sleep(time.Duration(attempt*100) * time.Millisecond)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				err := transferFn(ctx, source, remotePath)
				cancel()

				if err == nil {
					hr.Status = "success"
					hr.Changed = true
					hr.Size = srcSize
					if opts.Checksum {
						hr.Checksum = srcChecksum
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

// computeFileChecksum returns the SHA-256 hex digest of a file.
func computeFileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// isDir reports whether the given path looks like a directory
// (no file extension and ends without a dot-separated suffix).
// This is a heuristic used to decide if we need to append the basename.
func isDir(path string) bool {
	// If it ends with a slash, it is definitely a directory.
	if len(path) > 0 && path[len(path)-1] == '/' {
		return true
	}
	// If there is no dot in the last component, treat as directory.
	base := filepath.Base(path)
	return filepath.Ext(base) == ""
}
