// Package git_config provides Git configuration management.
package git_config

import (
	"fmt"
	"os/exec"
	"strings"
)

// ConfigResult represents the result of a git config operation.
type ConfigResult struct {
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// Get retrieves a git config value.
func Get(key, scope string) (ConfigResult, error) {
	if key == "" {
		return ConfigResult{Error: "key is required"}, fmt.Errorf("key is required")
	}

	args := []string{"config", "--get"}
	if scope != "" {
		args = append(args, "--"+scope)
	}
	args = append(args, key)

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Key not found
			return ConfigResult{Key: key, Value: ""}, nil
		}
		return ConfigResult{Key: key, Error: string(out)}, fmt.Errorf("git config --get failed: %w", err)
	}

	return ConfigResult{Key: key, Value: strings.TrimSpace(string(out))}, nil
}

// Set sets a git config value.
func Set(key, value, scope string) (ConfigResult, error) {
	if key == "" {
		return ConfigResult{Error: "key is required"}, fmt.Errorf("key is required")
	}

	// Check current value first
	current, err := Get(key, scope)
	if err != nil && current.Value != "" {
		return ConfigResult{Key: key, Error: err.Error()}, err
	}

	if current.Value == value {
		return ConfigResult{Key: key, Value: value, Changed: false}, nil
	}

	args := []string{"config"}
	if scope != "" {
		args = append(args, "--"+scope)
	}
	args = append(args, key, value)

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return ConfigResult{Key: key, Error: string(out)}, fmt.Errorf("git config failed: %w", err)
	}

	return ConfigResult{Key: key, Value: value, Changed: true}, nil
}

// Unset removes a git config value.
func Unset(key, scope string) (ConfigResult, error) {
	if key == "" {
		return ConfigResult{Error: "key is required"}, fmt.Errorf("key is required")
	}

	// Check if key exists
	current, err := Get(key, scope)
	if err != nil {
		return ConfigResult{Key: key, Error: err.Error()}, err
	}

	if current.Value == "" {
		return ConfigResult{Key: key, Changed: false}, nil
	}

	args := []string{"config", "--unset"}
	if scope != "" {
		args = append(args, "--"+scope)
	}
	args = append(args, key)

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return ConfigResult{Key: key, Error: string(out)}, fmt.Errorf("git config --unset failed: %w", err)
	}

	return ConfigResult{Key: key, Changed: true}, nil
}

// List returns all git config entries.
func List(scope string) (map[string]string, error) {
	args := []string{"config", "--list"}
	if scope != "" {
		args = append(args, "--"+scope)
	}

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git config --list failed: %w", err)
	}

	config := make(map[string]string)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			config[parts[0]] = parts[1]
		}
	}

	return config, nil
}
