// Package svn provides Subversion repository operations with structured results.
// Equivalent to Ansible's subversion module.
package svn

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by mutating operations.
type Result struct {
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Path     string `json:"path,omitempty"`
	Revision string `json:"revision,omitempty"`
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
}

// StatusEntry represents one entry from svn status.
type StatusEntry struct {
	Status string `json:"status"`
	Path   string `json:"path"`
}

// InfoResult represents the output of svn info.
type InfoResult struct {
	Status  string `json:"status"`
	Path    string `json:"path,omitempty"`
	URL     string `json:"url,omitempty"`
	Repo    string `json:"repo,omitempty"`
	Rev     string `json:"revision,omitempty"`
	Node    string `json:"node_kind,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// toolNotFound returns a consistent error when the svn binary is missing.
func toolNotFound(err error) error {
	return fmt.Errorf("svn command not found: %w", err)
}

// isNotFound checks whether an exec error indicates a missing binary.
func isNotFound(err error) bool {
	if _, ok := err.(*exec.Error); ok {
		return true
	}
	return false
}

// Checkout checks out a Subversion repository to dest.
// If dest already exists and force is false, returns Changed: false (idempotent).
func Checkout(url string, dest string, revision string, force bool) (Result, error) {
	if url == "" {
		return Result{Status: "failed", Error: "url is required"}, fmt.Errorf("url is required")
	}
	if dest == "" {
		return Result{Status: "failed", Error: "dest is required"}, fmt.Errorf("dest is required")
	}

	args := []string{"checkout", url, dest}
	if revision != "" {
		args = append(args, "--revision", revision)
	}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("svn", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if isNotFound(err) {
			return Result{Status: "failed", Error: "svn not found"}, toolNotFound(err)
		}
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("svn checkout: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Path: dest, Revision: revision, Output: output}, nil
}

// Update updates a working copy at dest to the given revision.
// Empty revision means HEAD.
func Update(dest string, revision string) (Result, error) {
	if dest == "" {
		return Result{Status: "failed", Error: "dest is required"}, fmt.Errorf("dest is required")
	}

	args := []string{"update", dest}
	if revision != "" {
		args = append(args, "--revision", revision)
	}

	cmd := exec.Command("svn", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if isNotFound(err) {
			return Result{Status: "failed", Error: "svn not found"}, toolNotFound(err)
		}
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("svn update: %v", err)}, err
	}

	changed := !strings.Contains(output, "At revision")
	return Result{Status: "success", Changed: changed, Path: dest, Revision: revision, Output: output}, nil
}

// Export exports a clean copy (without .svn metadata) from url to dest.
func Export(url string, dest string, revision string, force bool) (Result, error) {
	if url == "" {
		return Result{Status: "failed", Error: "url is required"}, fmt.Errorf("url is required")
	}
	if dest == "" {
		return Result{Status: "failed", Error: "dest is required"}, fmt.Errorf("dest is required")
	}

	args := []string{"export", url, dest}
	if revision != "" {
		args = append(args, "--revision", revision)
	}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("svn", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if isNotFound(err) {
			return Result{Status: "failed", Error: "svn not found"}, toolNotFound(err)
		}
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("svn export: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Path: dest, Revision: revision, Output: output}, nil
}

// Status returns the working copy status entries at dest.
func Status(dest string) ([]StatusEntry, error) {
	if dest == "" {
		return nil, fmt.Errorf("dest is required")
	}

	cmd := exec.Command("svn", "status", dest)
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return nil, toolNotFound(err)
		}
		return nil, fmt.Errorf("svn status: %w", err)
	}

	entries := make([]StatusEntry, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if len(line) < 8 {
			continue
		}
		status := strings.TrimSpace(line[:7])
		path := strings.TrimSpace(line[8:])
		if path != "" {
			entries = append(entries, StatusEntry{Status: status, Path: path})
		}
	}
	return entries, nil
}

// Info returns repository information for the working copy at dest.
func Info(dest string) (InfoResult, error) {
	if dest == "" {
		return InfoResult{Status: "failed", Error: "dest is required"}, fmt.Errorf("dest is required")
	}

	cmd := exec.Command("svn", "info", dest)
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return InfoResult{Status: "failed", Error: "svn not found"}, toolNotFound(err)
		}
		return InfoResult{Status: "failed", Error: fmt.Sprintf("svn info: %v", err)}, err
	}

	result := InfoResult{Status: "success", Path: dest, Output: strings.TrimSpace(string(out))}
	// Parse key fields from the output.
	for _, line := range strings.Split(result.Output, "\n") {
		switch {
		case strings.HasPrefix(line, "URL:"):
			result.URL = strings.TrimSpace(strings.TrimPrefix(line, "URL:"))
		case strings.HasPrefix(line, "Repository Root:"):
			result.Repo = strings.TrimSpace(strings.TrimPrefix(line, "Repository Root:"))
		case strings.HasPrefix(line, "Revision:"):
			result.Rev = strings.TrimSpace(strings.TrimPrefix(line, "Revision:"))
		case strings.HasPrefix(line, "Node Kind:"):
			result.Node = strings.TrimSpace(strings.TrimPrefix(line, "Node Kind:"))
		}
	}
	return result, nil
}

// Cleanup cleans up a working copy at dest.
func Cleanup(dest string) (Result, error) {
	if dest == "" {
		return Result{Status: "failed", Error: "dest is required"}, fmt.Errorf("dest is required")
	}

	cmd := exec.Command("svn", "cleanup", dest)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if isNotFound(err) {
			return Result{Status: "failed", Error: "svn not found"}, toolNotFound(err)
		}
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("svn cleanup: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Path: dest, Output: output}, nil
}

// Revert reverts local changes in a working copy at dest.
// If recursive is true, uses --recursive flag.
func Revert(dest string, recursive bool) (Result, error) {
	if dest == "" {
		return Result{Status: "failed", Error: "dest is required"}, fmt.Errorf("dest is required")
	}

	args := []string{"revert"}
	if recursive {
		args = append(args, "--recursive")
	}
	args = append(args, dest)

	cmd := exec.Command("svn", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if isNotFound(err) {
			return Result{Status: "failed", Error: "svn not found"}, toolNotFound(err)
		}
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("svn revert: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Path: dest, Output: output}, nil
}
