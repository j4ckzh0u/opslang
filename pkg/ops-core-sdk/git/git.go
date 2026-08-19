// Package git provides Git repository operations with structured results.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CloneResult represents the result of a git clone operation.
type CloneResult struct {
	Changed bool   `json:"changed"`
	Path    string `json:"path"`
	Error   string `json:"error,omitempty"`
}

// PullResult represents the result of a git pull operation.
type PullResult struct {
	Changed bool   `json:"changed"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// Clone clones a Git repository to the destination path.
// If the destination already exists, it returns Changed: false (idempotent).
// Supported options: branch, depth, bare (set to "true").
func Clone(url string, dest string, opts map[string]string) (CloneResult, error) {
	// Check if destination already exists (idempotent)
	if _, err := os.Stat(dest); err == nil {
		return CloneResult{Changed: false, Path: dest}, nil
	}

	// Build git clone arguments
	args := []string{"clone", url, dest}

	if opts != nil {
		if branch, ok := opts["branch"]; ok && branch != "" {
			args = append(args, "--branch", branch)
		}
		if depth, ok := opts["depth"]; ok && depth != "" {
			args = append(args, "--depth", depth)
		}
		if bare, ok := opts["bare"]; ok && bare == "true" {
			args = append(args, "--bare")
		}
	}

	// Execute git clone
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return CloneResult{
			Changed: false,
			Path:    dest,
			Error:   stderr.String(),
		}, fmt.Errorf("git clone failed: %w", err)
	}

	return CloneResult{Changed: true, Path: dest}, nil
}

// Pull performs a git pull on the repository at repoPath.
// If remote is empty, defaults to "origin".
// If branch is empty, pulls from the default branch.
// Returns Changed: false if already up to date.
func Pull(repoPath string, remote string, branch string) (PullResult, error) {
	// Default remote to origin
	if remote == "" {
		remote = "origin"
	}

	// Build git pull arguments
	args := []string{"pull", remote}
	if branch != "" {
		args = append(args, branch)
	}

	// Execute git pull
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return PullResult{
			Changed: false,
			Output:  stderr.String(),
			Error:   stderr.String(),
		}, fmt.Errorf("git pull failed: %w", err)
	}

	output := stdout.String()

	// Check if already up to date
	if strings.Contains(output, "Already up to date") {
		return PullResult{Changed: false, Output: output}, nil
	}

	return PullResult{Changed: true, Output: output}, nil
}
