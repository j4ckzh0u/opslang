// Package aptkey provides APT GPG key management.
package aptkey

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

// KeyInfo represents a GPG key.
type KeyInfo struct {
	ID   string `json:"id"`
	UID  string `json:"uid"`
	Size string `json:"size,omitempty"`
}

// ListResult is returned by key listing.
type ListResult struct {
	Keys  []KeyInfo `json:"keys"`
	Count int       `json:"count"`
	Error string    `json:"error,omitempty"`
}

func aptKey(args ...string) (string, error) {
	cmd := exec.Command("apt-key", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Add adds a GPG key from a URL.
func Add(url, keyring string) Result {
	if url == "" {
		return Result{Error: "URL is required"}
	}
	args := []string{"adv", "--fetch-keys", url}
	if keyring != "" {
		args = append(args, "--keyring", keyring)
	}
	out, err := aptKey(args...)
	if err != nil {
		return Result{Key: url, Error: fmt.Sprintf("apt-key add failed: %s: %s", err, out)}
	}
	return Result{Key: url, Success: true, Changed: true}
}

// AddFromKey adds a GPG key from a local file.
func AddFromKey(path, keyring string) Result {
	if path == "" {
		return Result{Error: "file path is required"}
	}
	args := []string{"add", path}
	if keyring != "" {
		args = append(args, "--keyring", keyring)
	}
	out, err := aptKey(args...)
	if err != nil {
		return Result{Key: path, Error: fmt.Sprintf("apt-key add failed: %s: %s", err, out)}
	}
	return Result{Key: path, Success: true, Changed: true}
}

// Remove removes a GPG key by ID.
func Remove(keyID, keyring string) Result {
	if keyID == "" {
		return Result{Error: "key ID is required"}
	}
	args := []string{"del", keyID}
	if keyring != "" {
		args = append(args, "--keyring", keyring)
	}
	out, err := aptKey(args...)
	if err != nil {
		return Result{Key: keyID, Error: fmt.Sprintf("apt-key del failed: %s: %s", err, out)}
	}
	return Result{Key: keyID, Success: true, Changed: true}
}

// List lists all trusted GPG keys.
func List() ListResult {
	out, err := aptKey("list")
	if err != nil {
		return ListResult{Error: fmt.Sprintf("apt-key list failed: %s: %s", err, out)}
	}
	var keys []KeyInfo
	var current KeyInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "pub") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				current = KeyInfo{
					Size: fields[1],
				}
				// Extract key ID from the pub line
				for _, f := range fields {
					if len(f) == 8 || len(f) == 16 {
						current.ID = f
					}
				}
			}
		} else if strings.HasPrefix(line, "uid") && current.ID != "" {
			current.UID = strings.TrimPrefix(line, "uid ")
			current.UID = strings.TrimSpace(current.UID)
			keys = append(keys, current)
			current = KeyInfo{}
		}
	}
	return ListResult{Keys: keys, Count: len(keys)}
}
