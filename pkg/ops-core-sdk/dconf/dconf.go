package dconf

import (
	"fmt"
	"os/exec"
	"strings"
)

// ReadResult represents the result of reading a dconf key.
type ReadResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Error string `json:"error,omitempty"`
}

// WriteResult represents the result of writing a dconf key.
type WriteResult struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ListResult represents the result of listing dconf keys.
type ListResult struct {
	Dir   string   `json:"dir"`
	Keys  []string `json:"keys"`
	Count int      `json:"count"`
	Error string   `json:"error,omitempty"`
}

// ResetResult represents the result of resetting a dconf key.
type ResetResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// Read reads the value of a dconf key.
func Read(key string) ReadResult {
	if key == "" {
		return ReadResult{Error: "key is required"}
	}

	cmd := exec.Command("dconf", "read", key)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ReadResult{
			Key:   key,
			Error: fmt.Sprintf("dconf read failed: %s: %s", err, string(out)),
		}
	}

	return ReadResult{
		Key:   key,
		Value: strings.TrimSpace(string(out)),
	}
}

// Write writes a value to a dconf key.
func Write(key, value string) WriteResult {
	if key == "" {
		return WriteResult{Error: "key is required"}
	}
	if value == "" {
		return WriteResult{Error: "value is required"}
	}

	// Read current value to check if changed
	current := Read(key)
	changed := current.Value != value

	cmd := exec.Command("dconf", "write", key, value)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return WriteResult{
			Key:   key,
			Value: value,
			Error: fmt.Sprintf("dconf write failed: %s: %s", err, string(out)),
		}
	}

	return WriteResult{
		Key:     key,
		Value:   value,
		Changed: changed,
	}
}

// List lists all keys in a dconf directory.
func List(dir string) ListResult {
	if dir == "" {
		dir = "/"
	}

	cmd := exec.Command("dconf", "list", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ListResult{
			Dir:   dir,
			Error: fmt.Sprintf("dconf list failed: %s: %s", err, string(out)),
		}
	}

	var keys []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			keys = append(keys, line)
		}
	}

	return ListResult{
		Dir:   dir,
		Keys:  keys,
		Count: len(keys),
	}
}

// Reset resets a dconf key to its default value.
func Reset(key string) ResetResult {
	if key == "" {
		return ResetResult{Error: "key is required"}
	}

	cmd := exec.Command("dconf", "reset", key)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ResetResult{
			Key:   key,
			Error: fmt.Sprintf("dconf reset failed: %s: %s", err, string(out)),
		}
	}

	return ResetResult{
		Key:     key,
		Changed: true,
	}
}
