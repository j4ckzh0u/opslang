// Package ip_netns provides network namespace management.
// Supports listing, creating, deleting, and executing commands in namespaces on Linux.
package ip_netns

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Namespace represents a network namespace.
type Namespace struct {
	Name   string `json:"name"`
	ID     string `json:"id,omitempty"`
	Exists bool   `json:"exists"`
}

// NamespaceResult represents the result of namespace operations.
type NamespaceResult struct {
	Success    bool        `json:"success"`
	Namespaces []Namespace `json:"namespaces,omitempty"`
	Changed    bool        `json:"changed,omitempty"`
	Output     string      `json:"output,omitempty"`
	Error      string      `json:"error,omitempty"`
	Duration   int64       `json:"duration_ms"`
}

// List returns all network namespaces.
func List() NamespaceResult {
	start := time.Now()

	out, err := exec.Command("ip", "netns", "list").Output()
	if err != nil {
		return NamespaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list namespaces: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	namespaces := parseNamespaces(string(out))
	return NamespaceResult{
		Success:    true,
		Namespaces: namespaces,
		Duration:   time.Since(start).Milliseconds(),
	}
}

// Get checks if a specific namespace exists.
func Get(name string) NamespaceResult {
	start := time.Now()

	// Try to get namespace info
	cmd := exec.Command("ip", "netns", "identify", name)
	err := cmd.Run()
	exists := err == nil

	return NamespaceResult{
		Success: true,
		Namespaces: []Namespace{
			{
				Name:   name,
				Exists: exists,
			},
		},
		Duration: time.Since(start).Milliseconds(),
	}
}

// Add creates a new network namespace.
func Add(name string) NamespaceResult {
	start := time.Now()

	cmd := exec.Command("ip", "netns", "add", name)
	if err := cmd.Run(); err != nil {
		return NamespaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create namespace %s: %v", name, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return NamespaceResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Delete removes a network namespace.
func Delete(name string) NamespaceResult {
	start := time.Now()

	cmd := exec.Command("ip", "netns", "delete", name)
	if err := cmd.Run(); err != nil {
		return NamespaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to delete namespace %s: %v", name, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return NamespaceResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Exec executes a command in a network namespace.
func Exec(namespace string, command string, args ...string) NamespaceResult {
	start := time.Now()

	// Build command: ip netns exec <namespace> <command> <args>
	cmdArgs := append([]string{"netns", "exec", namespace, command}, args...)
	cmd := exec.Command("ip", cmdArgs...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return NamespaceResult{
			Success:  false,
			Output:   string(out),
			Error:    fmt.Sprintf("command failed in namespace %s: %v", namespace, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return NamespaceResult{
		Success:  true,
		Output:   string(out),
		Duration: time.Since(start).Milliseconds(),
	}
}

// Pids returns the PIDs of processes in a namespace.
func Pids(name string) NamespaceResult {
	start := time.Now()

	out, err := exec.Command("ip", "netns", "pids", name).Output()
	if err != nil {
		return NamespaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to get PIDs for namespace %s: %v", name, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return NamespaceResult{
		Success:  true,
		Output:   string(out),
		Duration: time.Since(start).Milliseconds(),
	}
}

func parseNamespaces(output string) []Namespace {
	var namespaces []Namespace
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Format: "name (id: 0)" or just "name"
		ns := Namespace{
			Exists: true,
		}
		// Check if line contains "(id:"
		if idx := strings.Index(line, " (id: "); idx > 0 {
			ns.Name = line[:idx]
			// Extract ID
			idStart := idx + 6 // len(" (id: ")
			idEnd := strings.Index(line[idStart:], ")")
			if idEnd > 0 {
				ns.ID = line[idStart : idStart+idEnd]
			}
		} else {
			// No ID, just name
			ns.Name = strings.TrimSpace(line)
		}
		if ns.Name != "" {
			namespaces = append(namespaces, ns)
		}
	}
	return namespaces
}
