// Package hosts manages /etc/hosts entries for hostname resolution.
// Provides structured operations for adding, removing, and querying hosts entries.
package hosts

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Entry represents a hosts file entry.
type Entry struct {
	IP        string   `json:"ip"`
	Hostnames []string `json:"hostnames"`
	Comment   string   `json:"comment,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Entries []Entry `json:"entries"`
}

// ExistsResult is returned by Exists.
type ExistsResult struct {
	Exists bool `json:"exists"`
}

// AddResult is returned by Add.
type AddResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

const hostsFile = "/etc/hosts"

// List returns all entries from /etc/hosts.
func List() (ListResult, error) {
	f, err := os.Open(hostsFile)
	if err != nil {
		return ListResult{}, fmt.Errorf("open %s: %w", hostsFile, err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		entry := Entry{
			IP:        fields[0],
			Hostnames: fields[1:],
		}
		entries = append(entries, entry)
	}

	return ListResult{Entries: entries}, scanner.Err()
}

// Exists checks if a hostname exists in /etc/hosts.
func Exists(hostname string) (ExistsResult, error) {
	result, err := List()
	if err != nil {
		return ExistsResult{}, err
	}

	for _, entry := range result.Entries {
		for _, h := range entry.Hostnames {
			if h == hostname {
				return ExistsResult{Exists: true}, nil
			}
		}
	}

	return ExistsResult{Exists: false}, nil
}

// Add adds a hostname to /etc/hosts. If the hostname already exists, it updates the IP.
func Add(ip string, hostnames []string) (AddResult, error) {
	if ip == "" {
		return AddResult{Error: "IP address is required"}, fmt.Errorf("IP address is required")
	}
	if len(hostnames) == 0 {
		return AddResult{Error: "at least one hostname is required"}, fmt.Errorf("at least one hostname is required")
	}

	// Read current hosts
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return AddResult{Error: err.Error()}, fmt.Errorf("read %s: %w", hostsFile, err)
	}

	lines := strings.Split(string(content), "\n")
	hostnameSet := make(map[string]bool)
	for _, h := range hostnames {
		hostnameSet[h] = true
	}

	// Check if any of the hostnames already exist
	existingLine := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}

		// Check if any of our hostnames are in this line
		for _, h := range fields[1:] {
			if hostnameSet[h] {
				existingLine = i
				break
			}
		}
		if existingLine >= 0 {
			break
		}
	}

	// Build new entry line
	newLine := ip + "\t" + strings.Join(hostnames, " ")

	var changed bool
	if existingLine >= 0 {
		// Update existing line
		if lines[existingLine] != newLine {
			lines[existingLine] = newLine
			changed = true
		}
	} else {
		// Add new line
		lines = append(lines, newLine)
		changed = true
	}

	if !changed {
		return AddResult{Changed: false}, nil
	}

	// Write back
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(hostsFile, []byte(newContent), 0644); err != nil {
		return AddResult{Error: err.Error()}, fmt.Errorf("write %s: %w", hostsFile, err)
	}

	return AddResult{Changed: true}, nil
}

// Remove removes hostnames from /etc/hosts.
func Remove(hostnames []string) (RemoveResult, error) {
	if len(hostnames) == 0 {
		return RemoveResult{Error: "at least one hostname is required"}, fmt.Errorf("at least one hostname is required")
	}

	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return RemoveResult{Error: err.Error()}, fmt.Errorf("read %s: %w", hostsFile, err)
	}

	lines := strings.Split(string(content), "\n")
	hostnameSet := make(map[string]bool)
	for _, h := range hostnames {
		hostnameSet[h] = true
	}

	var newLines []string
	changed := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			newLines = append(newLines, line)
			continue
		}

		// Filter out hostnames to remove
		var keptHostnames []string
		for _, h := range fields[1:] {
			if !hostnameSet[h] {
				keptHostnames = append(keptHostnames, h)
			}
		}

		if len(keptHostnames) == 0 {
			// All hostnames removed, skip this line
			changed = true
			continue
		}

		if len(keptHostnames) < len(fields[1:]) {
			// Some hostnames removed
			changed = true
		}

		newLine := fields[0] + "\t" + strings.Join(keptHostnames, " ")
		newLines = append(newLines, newLine)
	}

	if !changed {
		return RemoveResult{Changed: false}, nil
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(hostsFile, []byte(newContent), 0644); err != nil {
		return RemoveResult{Error: err.Error()}, fmt.Errorf("write %s: %w", hostsFile, err)
	}

	return RemoveResult{Changed: true}, nil
}
