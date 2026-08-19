// Package acl provides POSIX ACL (Access Control List) management operations.
package acl

import (
	"fmt"
	"os/exec"
	"strings"
)

// ACLEntry represents a single ACL entry.
type ACLEntry struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
}

// GetResult represents the result of getting ACLs.
type GetResult struct {
	Path    string     `json:"path"`
	Entries []ACLEntry `json:"entries"`
}

// ActionResult represents the result of an ACL action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// Get returns the ACL entries for a file or directory.
func Get(path string) (*GetResult, error) {
	cmd := exec.Command("getfacl", "-p", "-c", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("getfacl failed: %w (output: %s)", err, string(out))
	}

	result := &GetResult{
		Path:    path,
		Entries: make([]ACLEntry, 0),
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			entry := ACLEntry{
				Type:       parts[0],
				Name:       parts[1],
				Permission: parts[2],
			}
			result.Entries = append(result.Entries, entry)
		}
	}
	return result, nil
}

// Set sets an ACL entry on a file or directory.
func Set(path string, entry string, recursive bool) (*ActionResult, error) {
	args := []string{"-m", entry}
	if recursive {
		args = append(args, "-R")
	}
	args = append(args, path)

	cmd := exec.Command("setfacl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("setfacl failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Set ACL %s on %s", entry, path)}, nil
}

// Remove removes an ACL entry from a file or directory.
func Remove(path string, entry string, recursive bool) (*ActionResult, error) {
	args := []string{"-x", entry}
	if recursive {
		args = append(args, "-R")
	}
	args = append(args, path)

	cmd := exec.Command("setfacl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("setfacl remove failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Removed ACL %s from %s", entry, path)}, nil
}

// RemoveAll removes all extended ACL entries from a file or directory.
func RemoveAll(path string, recursive bool) (*ActionResult, error) {
	args := []string{"-b"}
	if recursive {
		args = append(args, "-R")
	}
	args = append(args, path)

	cmd := exec.Command("setfacl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("setfacl remove-all failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Removed all ACLs from %s", path)}, nil
}
