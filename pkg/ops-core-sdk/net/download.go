// Download operations — Ansible get_url module equivalent.
package opsnet

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DownloadResult is returned by Download.
type DownloadResult struct {
	URL         string `json:"url"`
	Dest        string `json:"dest"`
	Size        int64  `json:"size"`
	Checksum    string `json:"checksum,omitempty"`
	StatusCode  int    `json:"status_code"`
	DurationMs  int64  `json:"duration_ms"`
	Changed     bool   `json:"changed"`
}

// Download downloads a URL to dest, optionally verifying checksum.
// checksumAlgo: "sha256", "sha1", "md5", or "" (no verify).
// checksumExpected: expected hex digest.
func Download(url string, dest string, checksumAlgo string, checksumExpected string) (DownloadResult, error) {
	result := DownloadResult{URL: url, Dest: dest}
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		return result, fmt.Errorf("net.Download: %w", err)
	}
	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("net.Download: HTTP %d from %s", resp.StatusCode, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return result, fmt.Errorf("net.Download: %w", err)
	}

	var writer io.Writer = out
	var h hash.Hash
	if checksumAlgo != "" {
		switch strings.ToLower(checksumAlgo) {
		case "md5":
			h = md5.New()
		case "sha1":
			h = sha1.New()
		case "sha256":
			h = sha256.New()
		default:
			out.Close()
			return result, fmt.Errorf("net.Download: unsupported checksum algorithm %q", checksumAlgo)
		}
		writer = io.MultiWriter(out, h)
	}

	size, err := io.Copy(writer, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(dest)
		return result, fmt.Errorf("net.Download: %w", err)
	}

	result.Size = size
	result.DurationMs = time.Since(start).Milliseconds()
	result.Changed = true

	if h != nil {
		result.Checksum = fmt.Sprintf("%x", h.Sum(nil))
		if checksumExpected != "" && result.Checksum != checksumExpected {
			os.Remove(dest)
			return result, fmt.Errorf("net.Download: checksum mismatch: got %s, expected %s", result.Checksum, checksumExpected)
		}
	}

	return result, nil
}
