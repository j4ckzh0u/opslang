// Package htpasswd manages htpasswd files for HTTP authentication.
// Equivalent to community.general.htpasswd module.
package htpasswd

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	File     string `json:"file,omitempty"`
	User     string `json:"user,omitempty"`
	Error    string `json:"error,omitempty"`
}

// InfoResult is returned by Info.
type InfoResult struct {
	Status string   `json:"status"`
	File   string   `json:"file"`
	Users  []string `json:"users"`
	Error  string   `json:"error,omitempty"`
}

// Set adds or updates a user in an htpasswd file using bcrypt (via htpasswd command).
func Set(path string, username string, password string, create bool) Result {
	if path == "" {
		return Result{Status: "failed", Error: "path is required"}
	}
	if username == "" {
		return Result{Status: "failed", Error: "username is required"}
	}
	if password == "" {
		return Result{Status: "failed", Error: "password is required"}
	}

	args := []string{"-nbB", username, password}
	if create {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			args = append([]string{"-cB", "-b", path, username, password}, args[3:]...)
			cmd := exec.Command("htpasswd", args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return Result{Status: "failed", Error: fmt.Sprintf("htpasswd: %v: %s", err, strings.TrimSpace(string(out)))}
			}
			return Result{Status: "success", Changed: true, File: path, User: username}
		}
	}

	cmd := exec.Command("htpasswd", "-nbB", username, password)
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("htpasswd: %v", err)}
	}

	newLine := strings.TrimSpace(string(out))

	// Read existing file
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new file
			if err := os.WriteFile(path, []byte(newLine+"\n"), 0644); err != nil {
				return Result{Status: "failed", Error: fmt.Sprintf("write: %v", err)}
			}
			return Result{Status: "success", Changed: true, File: path, User: username}
		}
		return Result{Status: "failed", Error: fmt.Sprintf("read: %v", err)}
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	found := false
	changed := false
	newLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, username+":") {
			found = true
			if line != newLine {
				changed = true
				newLines = append(newLines, newLine)
			} else {
				newLines = append(newLines, line)
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		newLines = append(newLines, newLine)
		changed = true
	}

	if changed {
		if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
			return Result{Status: "failed", Error: fmt.Sprintf("write: %v", err)}
		}
	}

	return Result{Status: "success", Changed: changed, File: path, User: username}
}

// Remove removes a user from an htpasswd file.
func Remove(path string, username string) Result {
	if path == "" || username == "" {
		return Result{Status: "failed", Error: "path and username are required"}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("read: %v", err)}
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	found := false
	newLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, username+":") {
			found = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !found {
		return Result{Status: "success", Changed: false, File: path, User: username}
	}

	if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("write: %v", err)}
	}

	return Result{Status: "success", Changed: true, File: path, User: username}
}

// Info returns users in an htpasswd file.
func Info(path string) InfoResult {
	if path == "" {
		return InfoResult{Status: "failed", Error: "path is required"}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return InfoResult{Status: "failed", Error: fmt.Sprintf("read: %v", err)}
	}

	users := make([]string, 0)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			users = append(users, parts[0])
		}
	}
	return InfoResult{Status: "success", File: path, Users: users}
}

// HashSHA1 generates a SHA1-hashed password (for basic auth {SHA} format).
func HashSHA1(password string) string {
	h := sha1.Sum([]byte(password))
	return "{SHA}" + base64.StdEncoding.EncodeToString(h[:])
}
