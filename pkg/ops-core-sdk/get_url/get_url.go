// Package get_url provides file download with checksum verification.
package get_url

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Result represents the result of a download operation.
type Result struct {
	URL      string `json:"url"`
	Dest     string `json:"dest"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum,omitempty"`
	Changed  bool   `json:"changed"`
	Message  string `json:"message"`
}

// Download downloads a file from url to dest with optional checksum verification.
// checksum format: "sha256:abc123..." or "md5:def456..."
func Download(url string, dest string, checksum string, force bool) (*Result, error) {
	result := &Result{URL: url, Dest: dest}

	// Check if file already exists with correct checksum
	if !force {
		if info, err := os.Stat(dest); err == nil {
			if checksum == "" {
				result.Changed = false
				result.Size = info.Size()
				result.Message = "file already exists"
				return result, nil
			}
			// Verify checksum
			actual, err := computeChecksum(dest, checksum)
			if err == nil && actual == extractHash(checksum) {
				result.Changed = false
				result.Size = info.Size()
				result.Checksum = actual
				result.Message = "file exists with correct checksum"
				return result, nil
			}
		}
	}

	// Download file
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get_url.Download: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get_url.Download: HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return nil, fmt.Errorf("get_url.Download: failed to create directory: %w", err)
	}

	// Write to temp file first
	tmpDest := dest + ".tmp"
	out, err := os.Create(tmpDest)
	if err != nil {
		return nil, fmt.Errorf("get_url.Download: failed to create file: %w", err)
	}

	var writer io.Writer = out
	var hashWriter hash.Hash
	if checksum != "" {
		hashWriter = newHasher(checksum)
		writer = io.MultiWriter(out, hashWriter)
	}

	size, err := io.Copy(writer, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tmpDest)
		return nil, fmt.Errorf("get_url.Download: failed to write file: %w", err)
	}

	// Verify checksum if provided
	if checksum != "" && hashWriter != nil {
		actual := hex.EncodeToString(hashWriter.Sum(nil))
		expected := extractHash(checksum)
		if actual != expected {
			os.Remove(tmpDest)
			return nil, fmt.Errorf("get_url.Download: checksum mismatch: expected %s, got %s", expected, actual)
		}
		result.Checksum = actual
	}

	// Rename temp to final
	if err := os.Rename(tmpDest, dest); err != nil {
		return nil, fmt.Errorf("get_url.Download: failed to rename file: %w", err)
	}

	result.Size = size
	result.Changed = true
	result.Message = "downloaded successfully"
	return result, nil
}

func extractHash(checksum string) string {
	parts := strings.SplitN(checksum, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return checksum
}

func extractAlgo(checksum string) string {
	parts := strings.SplitN(checksum, ":", 2)
	if len(parts) == 2 {
		return strings.ToLower(parts[0])
	}
	return "sha256"
}

func computeChecksum(path string, checksum string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := newHasher(checksum)
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func newHasher(checksum string) hash.Hash {
	switch extractAlgo(checksum) {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	default:
		return sha256.New()
	}
}

// multiWriter writes to multiple writers
type multiWriter struct {
	writers []io.Writer
}

func (mw *multiWriter) Write(p []byte) (int, error) {
	for _, w := range mw.writers {
		n, err := w.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			return n, io.ErrShortWrite
		}
	}
	return len(p), nil
}
