// Package sefcontext manages SELinux file context mappings.
// Requires SELinux to be enabled and semanage to be installed.
package sefcontext

import (
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a file context operation.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
}

// ContextEntry represents a SELinux file context mapping.
type ContextEntry struct {
	SELinuxContext string `json:"selinux_context"`
	FileSpec       string `json:"filespec"`
}

// ListResult represents the result of listing file context mappings.
type ListResult struct {
	Contexts []ContextEntry `json:"contexts"`
}

// Add adds a file context mapping for the given file specification.
func Add(filespec, seType string) (*ActionResult, error) {
	if filespec == "" {
		return nil, fmt.Errorf("filespec is required")
	}
	if seType == "" {
		return nil, fmt.Errorf("se_type is required")
	}

	// Check if already exists
	context := fmt.Sprintf("unconfined_u:object_r:%s:s0", seType)
	entries, err := listEntries()
	if err == nil {
		for _, e := range entries {
			if e.FileSpec == filespec && e.SELinuxContext == context {
				return &ActionResult{Changed: false, Message: "file context already exists"}, nil
			}
		}
	}

	cmd := exec.Command("semanage", "fcontext", "-a", "-t", seType, filespec)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to add file context: %s: %w", string(out), err)
	}
	return &ActionResult{Changed: true, Message: "file context added"}, nil
}

// Modify modifies an existing file context mapping.
func Modify(filespec, seType string) (*ActionResult, error) {
	if filespec == "" {
		return nil, fmt.Errorf("filespec is required")
	}
	if seType == "" {
		return nil, fmt.Errorf("se_type is required")
	}

	cmd := exec.Command("semanage", "fcontext", "-m", "-t", seType, filespec)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to modify file context: %s: %w", string(out), err)
	}
	return &ActionResult{Changed: true, Message: "file context modified"}, nil
}

// Remove removes a file context mapping.
func Remove(filespec string) (*ActionResult, error) {
	if filespec == "" {
		return nil, fmt.Errorf("filespec is required")
	}

	cmd := exec.Command("semanage", "fcontext", "-d", filespec)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to remove file context: %s: %w", string(out), err)
	}
	return &ActionResult{Changed: true, Message: "file context removed"}, nil
}

// List returns all SELinux file context mappings.
func List() (*ListResult, error) {
	entries, err := listEntries()
	if err != nil {
		return nil, err
	}
	return &ListResult{Contexts: entries}, nil
}

// Apply applies file context mappings to files matching the specification (runs restorecon).
func Apply(filespec string, recursive bool) (*ActionResult, error) {
	if filespec == "" {
		return nil, fmt.Errorf("filespec is required")
	}

	args := []string{}
	if recursive {
		args = append(args, "-R")
	}
	args = append(args, filespec)

	cmd := exec.Command("restorecon", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to apply file context: %s: %w", string(out), err)
	}
	return &ActionResult{Changed: true, Message: "file context applied"}, nil
}

func listEntries() ([]ContextEntry, error) {
	cmd := exec.Command("semanage", "fcontext", "-l")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list file contexts: %s: %w", string(out), err)
	}

	entries := make([]ContextEntry, 0)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "SELinux") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		entries = append(entries, ContextEntry{
			FileSpec:       fields[0],
			SELinuxContext: fields[len(fields)-1],
		})
	}
	return entries, nil
}
