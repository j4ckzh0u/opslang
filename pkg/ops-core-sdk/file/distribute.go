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
	"strings"
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
	Checksum bool          `json:"checksum"` // verify remote content hash after transfer
	Mode     string        `json:"mode"`     // optional octal mode applied remotely, e.g. "0644"
	Parallel int           `json:"parallel"` // max concurrent transfers
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"` // total attempts per host
}

// VerifyFunc compares the remote file against an expected SHA-256 digest.
type VerifyFunc func(ctx context.Context, endpoint, wantSHA256 string) error

// ChmodFunc applies a permission mode to the remote file.
type ChmodFunc func(ctx context.Context, endpoint, mode string) error

// DefaultVerifyFunc and DefaultChmodFunc are the remote-side hooks; the
// package wires real SFTP implementations (see ssh_transfer.go).
var DefaultVerifyFunc VerifyFunc = func(_ context.Context, _, _ string) error {
	return fmt.Errorf("no verify function configured; set file.DefaultVerifyFunc")
}

var DefaultChmodFunc ChmodFunc = func(_ context.Context, _, _ string) error {
	return fmt.Errorf("no chmod function configured; set file.DefaultChmodFunc")
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

// DefaultTransferFunc is used by Distribute when no explicit transfer
// function is provided. The package wires a real SSH/SFTP implementation
// (see ssh_transfer.go); tests may override it.
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

			// Determine the remote path: Dest is used verbatim; a trailing
			// "/" means "directory, keep the source basename". A path with
			// no extension is NOT guessed to be a directory any more -
			// guessing sent /usr/local/bin/mytool to .../mytool/mytool.
			remotePath := t.Dest
			if remotePath == "" {
				remotePath = "/tmp/" + filepath.Base(source)
			} else if strings.HasSuffix(remotePath, "/") {
				remotePath += filepath.Base(source)
			}

			user := t.User
			if user == "" {
				user = "root"
			}
			port := t.Port
			if port == 0 {
				port = 22
			}
			endpoint := formatEndpoint(user, t.Host, port, remotePath)

			var lastErr error
			for attempt := 0; attempt < retries; attempt++ {
				if attempt > 0 {
					time.Sleep(time.Duration(attempt*100) * time.Millisecond)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				err := transferFn(ctx, source, endpoint)
				if err == nil && opts.Checksum && srcChecksum != "" {
					// Verify the remote content really matches: a reported
					// "checksum" copied from the local file proved nothing.
					err = DefaultVerifyFunc(ctx, endpoint, srcChecksum)
				}
				if err == nil && opts.Mode != "" {
					err = DefaultChmodFunc(ctx, endpoint, opts.Mode)
				}
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

