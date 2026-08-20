// Package sys_persist provides persistent kernel parameter management via /etc/sysctl.d/.
package sys_persist

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ActionResult represents the result of a sysctl persistence action.
type ActionResult struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Changed    bool   `json:"changed"`
	Persisted  bool   `json:"persisted"`
	FilePath   string `json:"file_path,omitempty"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// GetResult represents the current persisted value.
type GetResult struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	FilePath string `json:"file_path,omitempty"`
	Found    bool   `json:"found"`
}

// ListResult represents all persisted sysctl settings.
type ListResult struct {
	Settings []PersistedSetting `json:"settings"`
	Count    int                `json:"count"`
}

// PersistedSetting represents a single persisted sysctl setting.
type PersistedSetting struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	FilePath string `json:"file_path"`
}

const sysctlDir = "/etc/sysctl.d"

// Set persists a sysctl parameter to /etc/sysctl.d/99-opslang.conf (idempotent).
func Set(name, value string) (*ActionResult, error) {
	if name == "" {
		return nil, fmt.Errorf("parameter name is required")
	}
	if value == "" {
		return nil, fmt.Errorf("parameter value is required")
	}

	start := time.Now()
	confPath := filepath.Join(sysctlDir, "99-opslang.conf")

	// Ensure directory exists
	if err := os.MkdirAll(sysctlDir, 0755); err != nil {
		return nil, fmt.Errorf("create sysctl.d directory: %w", err)
	}

	// Read existing file
	existing := ""
	if data, err := os.ReadFile(confPath); err == nil {
		existing = string(data)
	}

	prefix := name + " = "
	newLine := fmt.Sprintf("%s%s\n", prefix, value)

	// Check if already set correctly
	for _, line := range strings.Split(existing, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			currentVal := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			if currentVal == value {
				return &ActionResult{
					Name:       name,
					Value:      value,
					Changed:    false,
					Persisted:  true,
					FilePath:   confPath,
					DurationMs: time.Since(start).Milliseconds(),
				}, nil
			}
		}
	}

	// Replace or append
	var newContent strings.Builder
	replaced := false
	for _, line := range strings.Split(existing, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			newContent.WriteString(newLine)
			replaced = true
		} else if trimmed != "" || newContent.Len() > 0 {
			newContent.WriteString(line + "\n")
		}
	}
	if !replaced {
		newContent.WriteString(newLine)
	}

	if err := os.WriteFile(confPath, []byte(newContent.String()), 0644); err != nil {
		return nil, fmt.Errorf("write sysctl config: %w", err)
	}

	// Apply immediately
	if err := exec.Command("sysctl", "-p", confPath).Run(); err != nil {
		return &ActionResult{
			Name:       name,
			Value:      value,
			Changed:    true,
			Persisted:  true,
			FilePath:   confPath,
			DurationMs: time.Since(start).Milliseconds(),
			Error:      fmt.Sprintf("persisted but sysctl -p failed: %v", err),
		}, nil
	}

	return &ActionResult{
		Name:       name,
		Value:      value,
		Changed:    true,
		Persisted:  true,
		FilePath:   confPath,
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Get returns the current persisted value of a sysctl parameter.
func Get(name string) (*GetResult, error) {
	if name == "" {
		return nil, fmt.Errorf("parameter name is required")
	}

	result := &GetResult{Name: name}
	prefix := name + " = "

	// Search all sysctl.d files
	entries, err := os.ReadDir(sysctlDir)
	if err != nil {
		return result, nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}
		fpath := filepath.Join(sysctlDir, entry.Name())
		data, err := os.ReadFile(fpath)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, prefix) {
				result.Value = strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
				result.FilePath = fpath
				result.Found = true
				return result, nil
			}
		}
	}

	return result, nil
}

// Remove removes a persisted sysctl parameter (idempotent).
func Remove(name string) (*ActionResult, error) {
	if name == "" {
		return nil, fmt.Errorf("parameter name is required")
	}

	start := time.Now()
	confPath := filepath.Join(sysctlDir, "99-opslang.conf")
	prefix := name + " = "

	data, err := os.ReadFile(confPath)
	if err != nil {
		return &ActionResult{
			Name:       name,
			Changed:    false,
			DurationMs: time.Since(start).Milliseconds(),
		}, nil
	}

	var newContent strings.Builder
	found := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			found = true
			continue
		}
		if trimmed != "" || newContent.Len() > 0 {
			newContent.WriteString(line + "\n")
		}
	}

	if !found {
		return &ActionResult{
			Name:       name,
			Changed:    false,
			DurationMs: time.Since(start).Milliseconds(),
		}, nil
	}

	if err := os.WriteFile(confPath, []byte(newContent.String()), 0644); err != nil {
		return nil, fmt.Errorf("write sysctl config: %w", err)
	}

	return &ActionResult{
		Name:       name,
		Changed:    true,
		Persisted:  false,
		FilePath:   confPath,
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// List returns all persisted sysctl settings from /etc/sysctl.d/.
func List() (*ListResult, error) {
	result := &ListResult{Settings: make([]PersistedSetting, 0)}

	entries, err := os.ReadDir(sysctlDir)
	if err != nil {
		return result, nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}
		fpath := filepath.Join(sysctlDir, entry.Name())
		data, err := os.ReadFile(fpath)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
				continue
			}
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) != 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result.Settings = append(result.Settings, PersistedSetting{
				Name:     name,
				Value:    value,
				FilePath: fpath,
			})
		}
	}

	result.Count = len(result.Settings)
	return result, nil
}
