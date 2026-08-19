// Package crypttab provides /etc/crypttab management operations.
package crypttab

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ActionResult represents the result of a crypttab operation.
type ActionResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Entry represents a crypttab entry.
type Entry struct {
	Name       string `json:"name"`
	Device     string `json:"device"`
	KeyFile    string `json:"key_file"`
	Options    string `json:"options"`
}

// ListResult represents the result of listing crypttab entries.
type ListResult struct {
	Entries []Entry `json:"entries"`
}

// Add adds an entry to /etc/crypttab.
func Add(name string, device string, keyFile string, options string) (ActionResult, error) {
	if name == "" || device == "" {
		return ActionResult{}, fmt.Errorf("name and device are required")
	}

	crypttabPath := "/etc/crypttab"
	content, err := os.ReadFile(crypttabPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return ActionResult{Name: name, Success: false, Error: err.Error()}, fmt.Errorf("failed to read crypttab: %w", err)
		}
		content = []byte{}
	}

	// Check if entry already exists
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return ActionResult{Name: name, Changed: false, Success: true}, nil
		}
	}

	// Append new entry
	entry := fmt.Sprintf("%s %s %s %s\n", name, device, keyFile, options)
	newContent := string(content) + entry

	if err := os.WriteFile(crypttabPath, []byte(newContent), 0644); err != nil {
		return ActionResult{Name: name, Success: false, Error: err.Error()}, fmt.Errorf("failed to write crypttab: %w", err)
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Remove removes an entry from /etc/crypttab.
func Remove(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("name is required")
	}

	crypttabPath := "/etc/crypttab"
	content, err := os.ReadFile(crypttabPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ActionResult{Name: name, Changed: false, Success: true}, nil
		}
		return ActionResult{Name: name, Success: false, Error: err.Error()}, fmt.Errorf("failed to read crypttab: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	found := false

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			found = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !found {
		return ActionResult{Name: name, Changed: false, Success: true}, nil
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(crypttabPath, []byte(newContent), 0644); err != nil {
		return ActionResult{Name: name, Success: false, Error: err.Error()}, fmt.Errorf("failed to write crypttab: %w", err)
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Modify modifies an existing entry in /etc/crypttab.
func Modify(name string, device string, keyFile string, options string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("name is required")
	}

	crypttabPath := "/etc/crypttab"
	content, err := os.ReadFile(crypttabPath)
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: err.Error()}, fmt.Errorf("failed to read crypttab: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false

	for i, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			found = true
			newEntry := fmt.Sprintf("%s %s %s %s", name, device, keyFile, options)
			lines[i] = newEntry
			break
		}
	}

	if !found {
		return ActionResult{Name: name, Success: false, Error: "entry not found"}, fmt.Errorf("entry %s not found", name)
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(crypttabPath, []byte(newContent), 0644); err != nil {
		return ActionResult{Name: name, Success: false, Error: err.Error()}, fmt.Errorf("failed to write crypttab: %w", err)
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Get retrieves a specific entry from /etc/crypttab.
func Get(name string) (Entry, error) {
	if name == "" {
		return Entry{}, fmt.Errorf("name is required")
	}

	crypttabPath := "/etc/crypttab"
	content, err := os.ReadFile(crypttabPath)
	if err != nil {
		return Entry{}, fmt.Errorf("failed to read crypttab: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			entry := Entry{Name: fields[0]}
			if len(fields) > 1 {
				entry.Device = fields[1]
			}
			if len(fields) > 2 {
				entry.KeyFile = fields[2]
			}
			if len(fields) > 3 {
				entry.Options = strings.Join(fields[3:], " ")
			}
			return entry, nil
		}
	}

	return Entry{}, fmt.Errorf("entry %s not found", name)
}

// List lists all entries in /etc/crypttab.
func List() (ListResult, error) {
	crypttabPath := "/etc/crypttab"
	content, err := os.ReadFile(crypttabPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ListResult{Entries: make([]Entry, 0)}, nil
		}
		return ListResult{}, fmt.Errorf("failed to read crypttab: %w", err)
	}

	result := ListResult{Entries: make([]Entry, 0)}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			entry := Entry{Name: fields[0]}
			if len(fields) > 1 {
				entry.Device = fields[1]
			}
			if len(fields) > 2 {
				entry.KeyFile = fields[2]
			}
			if len(fields) > 3 {
				entry.Options = strings.Join(fields[3:], " ")
			}
			result.Entries = append(result.Entries, entry)
		}
	}

	return result, nil
}

// Exists checks if an entry exists in /etc/crypttab.
func Exists(name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name is required")
	}

	crypttabPath := "/etc/crypttab"
	content, err := os.ReadFile(crypttabPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read crypttab: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return true, nil
		}
	}

	return false, nil
}

// Validate validates the crypttab file syntax.
func Validate() (map[string]interface{}, error) {
	crypttabPath := "/etc/crypttab"
	content, err := os.ReadFile(crypttabPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{
				"valid":   true,
				"errors":  []string{},
				"entries": 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to read crypttab: %w", err)
	}

	var errors []string
	entryCount := 0

	lines := strings.Split(string(content), "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entryCount++
		fields := strings.Fields(line)
		if len(fields) < 2 {
			errors = append(errors, fmt.Sprintf("line %d: insufficient fields", lineNum+1))
		}
	}

	return map[string]interface{}{
		"valid":   len(errors) == 0,
		"errors":  errors,
		"entries": entryCount,
	}, nil
}

// Backup creates a backup of /etc/crypttab.
func Backup(backupDir string) (map[string]interface{}, error) {
	crypttabPath := "/etc/crypttab"
	if backupDir == "" {
		backupDir = "/tmp"
	}

	// Check if crypttab exists
	if _, err := os.Stat(crypttabPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"success":  false,
			"backup":   "",
			"error":    "crypttab does not exist",
		}, nil
	}

	// Create backup
	content, err := os.ReadFile(crypttabPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read crypttab: %w", err)
	}

	backupPath := filepath.Join(backupDir, "crypttab.backup")
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write backup: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"backup":  backupPath,
	}, nil
}
