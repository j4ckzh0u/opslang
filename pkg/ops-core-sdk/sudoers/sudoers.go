// Package sudoers manages sudoers configuration.
// Equivalent to community.general.sudoers module.
package sudoers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Name    string `json:"name,omitempty"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

// InfoResult is returned by Info.
type InfoResult struct {
	Status  string `json:"status"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Exists  bool   `json:"exists"`
	Error   string `json:"error,omitempty"`
}

// Set creates or updates a sudoers drop-in file.
func Set(name string, user string, commands string, nopasswd bool, sudoersDir string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}
	if user == "" {
		return Result{Status: "failed", Error: "user is required"}
	}
	if commands == "" {
		commands = "ALL"
	}
	if sudoersDir == "" {
		sudoersDir = "/etc/sudoers.d"
	}

	nopasswdStr := ""
	if nopasswd {
		nopasswdStr = "NOPASSWD: "
	}

	content := fmt.Sprintf("# Managed by OpsLang\n%s %s= %s%s\n", user, "ALL=(ALL)", nopasswdStr, commands)

	path := filepath.Join(sudoersDir, name)
	existing, err := os.ReadFile(path)
	if err == nil && strings.TrimSpace(string(existing)) == strings.TrimSpace(content) {
		return Result{Status: "success", Changed: false, Name: name, Path: path}
	}

	if err := os.MkdirAll(sudoersDir, 0750); err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("mkdir: %v", err)}
	}

	if err := os.WriteFile(path, []byte(content), 0440); err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("write: %v", err)}
	}

	return Result{Status: "success", Changed: true, Name: name, Path: path}
}

// Remove removes a sudoers drop-in file.
func Remove(name string, sudoersDir string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}
	if sudoersDir == "" {
		sudoersDir = "/etc/sudoers.d"
	}

	path := filepath.Join(sudoersDir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Result{Status: "success", Changed: false, Name: name, Path: path}
	}

	if err := os.Remove(path); err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("remove: %v", err)}
	}
	return Result{Status: "success", Changed: true, Name: name, Path: path}
}

// Info returns information about a sudoers drop-in file.
func Info(name string, sudoersDir string) InfoResult {
	if name == "" {
		return InfoResult{Status: "failed", Error: "name is required"}
	}
	if sudoersDir == "" {
		sudoersDir = "/etc/sudoers.d"
	}

	path := filepath.Join(sudoersDir, name)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return InfoResult{Status: "success", Name: name, Path: path, Exists: false}
		}
		return InfoResult{Status: "failed", Error: fmt.Sprintf("read: %v", err)}
	}

	return InfoResult{Status: "success", Name: name, Path: path, Exists: true, Content: strings.TrimSpace(string(content))}
}
