// Package fetch provides remote file fetching operations.
package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Result represents the result of a fetch operation.
type Result struct {
	Source   string `json:"source"`
	Dest     string `json:"dest"`
	Size     int64  `json:"size"`
	Changed  bool   `json:"changed"`
}

// File copies a file from a source path to a destination path.
func File(source string, dest string) (*Result, error) {
	src, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return nil, fmt.Errorf("failed to create dest directory: %w", err)
	}

	dst, err := os.Create(dest)
	if err != nil {
		return nil, fmt.Errorf("failed to create dest file: %w", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy: %w", err)
	}

	return &Result{
		Source:  source,
		Dest:    dest,
		Size:    size,
		Changed: true,
	}, nil
}

// URL downloads a file from a URL to a local path.
func URL(url string, dest string) (*Result, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return nil, fmt.Errorf("failed to create dest directory: %w", err)
	}

	dst, err := os.Create(dest)
	if err != nil {
		return nil, fmt.Errorf("failed to create dest file: %w", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to save: %w", err)
	}

	return &Result{
		Source:  url,
		Dest:    dest,
		Size:    size,
		Changed: true,
	}, nil
}
