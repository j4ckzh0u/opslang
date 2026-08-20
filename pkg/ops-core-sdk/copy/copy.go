// Package copy provides Ansible copy module equivalent.
package copy

import (
	"io"
	"os"
	"path/filepath"
)

// CopyResult is returned by Copy.
type CopyResult struct {
	Dest    string `json:"dest"`
	Src     string `json:"src,omitempty"`
	Changed bool   `json:"changed"`
	Mode    string `json:"mode,omitempty"`
	Owner   string `json:"owner,omitempty"`
	Checksum string `json:"checksum,omitempty"`
	Error   string `json:"error,omitempty"`
}

// File copies src to dest.
func File(src, dest, mode, owner, group string, backup bool) CopyResult {
	if src == "" {
		return CopyResult{Error: "src is required"}
	}
	if dest == "" {
		return CopyResult{Error: "dest is required"}
	}

	info, err := os.Stat(src)
	if err != nil {
		return CopyResult{Error: err.Error()}
	}
	if info.IsDir() {
		return CopyResult{Error: "src is a directory; use synchronize instead"}
	}

	// Check if dest exists with same content
	if existing, err := os.Stat(dest); err == nil {
		if !existing.IsDir() {
			srcData, _ := os.ReadFile(src)
			dstData, _ := os.ReadFile(dest)
			if string(srcData) == string(dstData) {
				return CopyResult{Dest: dest, Src: src, Changed: false}
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return CopyResult{Error: err.Error()}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return CopyResult{Error: err.Error()}
	}
	defer srcFile.Close()

	perm := os.FileMode(0644)
	if mode != "" {
		if m, err := parseMode(mode); err == nil {
			perm = m
		}
	}

	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return CopyResult{Error: err.Error()}
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return CopyResult{Error: err.Error()}
	}

	return CopyResult{Dest: dest, Src: src, Changed: true, Mode: mode, Owner: owner}
}

// Content writes content directly to dest.
func Content(content, dest, mode, owner, group string, backup bool) CopyResult {
	if dest == "" {
		return CopyResult{Error: "dest is required"}
	}

	if existing, err := os.ReadFile(dest); err == nil {
		if string(existing) == content {
			return CopyResult{Dest: dest, Changed: false}
		}
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return CopyResult{Error: err.Error()}
	}

	perm := os.FileMode(0644)
	if mode != "" {
		if m, err := parseMode(mode); err == nil {
			perm = m
		}
	}

	if err := os.WriteFile(dest, []byte(content), perm); err != nil {
		return CopyResult{Error: err.Error()}
	}
	return CopyResult{Dest: dest, Changed: true, Mode: mode, Owner: owner}
}

func parseMode(s string) (os.FileMode, error) {
	var mode uint32
	for _, c := range s {
		if c < '0' || c > '7' {
			return 0, os.ErrInvalid
		}
		mode = mode*8 + uint32(c-'0')
	}
	return os.FileMode(mode), nil
}
