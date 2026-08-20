// Package tempfile provides Ansible tempfile module equivalent.
package tempfile

import (
	"os"
	"path/filepath"
)

// TempfileResult contains the created temp file/dir path.
type TempfileResult struct {
	Path    string `json:"path"`
	Changed bool   `json:"changed"`
	State   string `json:"state"` // "file" or "directory"
	Error   string `json:"error,omitempty"`
}

// CreateFile creates a temporary file.
func CreateFile(prefix, suffix, path string) TempfileResult {
	if prefix == "" {
		prefix = "ansible-"
	}
	if path == "" {
		path = os.TempDir()
	}
	f, err := os.CreateTemp(path, prefix+"*"+suffix)
	if err != nil {
		return TempfileResult{Error: err.Error()}
	}
	_ = f.Close()
	return TempfileResult{Path: f.Name(), Changed: true, State: "file"}
}

// CreateDir creates a temporary directory.
func CreateDir(prefix, suffix, path string) TempfileResult {
	if prefix == "" {
		prefix = "ansible-"
	}
	if path == "" {
		path = os.TempDir()
	}
	d, err := os.MkdirTemp(path, prefix+"*"+suffix)
	if err != nil {
		return TempfileResult{Error: err.Error()}
	}
	return TempfileResult{Path: d, Changed: true, State: "directory"}
}

// Delete removes a temp file or directory.
func Delete(path string) TempfileResult {
	if path == "" {
		return TempfileResult{Error: "path is required"}
	}
	if !filepath.IsAbs(path) {
		return TempfileResult{Error: "path must be absolute"}
	}
	if err := os.RemoveAll(path); err != nil {
		return TempfileResult{Error: err.Error()}
	}
	return TempfileResult{Path: path, Changed: true}
}
