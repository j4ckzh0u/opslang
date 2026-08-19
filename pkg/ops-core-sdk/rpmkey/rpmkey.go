// Package rpmkey provides RPM GPG key management.
package rpmkey

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by key operations.
type Result struct {
	Key     string `json:"key,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by key listing.
type ListResult struct {
	Keys []string `json:"keys"`
	Count int     `json:"count"`
	Error string  `json:"error,omitempty"`
}

func rpm(args ...string) (string, error) {
	cmd := exec.Command("rpm", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Import imports a GPG key.
func Import(keyPath string) Result {
	if keyPath == "" {
		return Result{Error: "key path is required"}
	}
	out, err := rpm("--import", keyPath)
	if err != nil {
		return Result{Key: keyPath, Error: fmt.Sprintf("rpm import failed: %s: %s", err, out)}
	}
	return Result{Key: keyPath, Success: true, Changed: true}
}

// List lists installed GPG keys.
func List() ListResult {
	out, err := rpm("-qa", "gpg-pubkey*")
	if err != nil {
		return ListResult{Error: fmt.Sprintf("rpm list keys failed: %s: %s", err, out)}
	}
	var keys []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			keys = append(keys, line)
		}
	}
	return ListResult{Keys: keys, Count: len(keys)}
}

// Remove removes a GPG key by key ID.
func Remove(keyID string) Result {
	if keyID == "" {
		return Result{Error: "key ID is required"}
	}
	out, err := rpm("--erase", keyID)
	if err != nil {
		return Result{Key: keyID, Error: fmt.Sprintf("rpm erase key failed: %s: %s", err, out)}
	}
	return Result{Key: keyID, Success: true, Changed: true}
}
