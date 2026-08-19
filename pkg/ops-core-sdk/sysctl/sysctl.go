// Package sysctl provides structured Linux kernel parameter operations for OpsLang.
// All functions read from and write to /proc/sys directly, never through shell calls.
// Results are strongly-typed structs with JSON serialization support.
package sysctl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// procSysRoot is the root directory for kernel parameters.
// It is a variable so tests can override it.
var procSysRoot = "/proc/sys"

// SysctlResult represents a single kernel parameter name-value pair.
type SysctlResult struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetResult is returned by Get, holding the current and default values.
type GetResult struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Default string `json:"default"`
}

// SetResult is returned by Set, reporting whether the value was changed.
type SetResult struct {
	Changed bool   `json:"changed"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Error   string `json:"error,omitempty"`
}

// nameToPath converts a dotted sysctl name (e.g. "net.ipv4.ip_forward")
// to its /proc/sys file path (e.g. "/proc/sys/net/ipv4/ip_forward").
func nameToPath(name string) string {
	return filepath.Join(procSysRoot, filepath.FromSlash(strings.ReplaceAll(name, ".", "/")))
}

// pathToName converts a /proc/sys file path back to dotted notation.
func pathToName(path string) string {
	rel, err := filepath.Rel(procSysRoot, path)
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(filepath.ToSlash(rel), "/", ".")
}

// readValue reads and trims the content of a sysctl parameter file.
func readValue(name string) (string, error) {
	path := nameToPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// Get reads the current value of a kernel parameter.
// The name uses dotted notation (e.g. "net.ipv4.ip_forward").
// The Default field is left empty as Linux does not expose persistent defaults.
func Get(name string) (GetResult, error) {
	value, err := readValue(name)
	if err != nil {
		return GetResult{}, fmt.Errorf("sysctl.Get: failed to read %q: %w", name, err)
	}
	return GetResult{
		Name:  name,
		Value: value,
	}, nil
}

// Set writes a value to a kernel parameter if it differs from the current value.
// Returns Changed=true only when the value was actually written.
func Set(name string, value string) (SetResult, error) {
	result := SetResult{
		Name:  name,
		Value: value,
	}

	current, err := readValue(name)
	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("sysctl.Set: failed to read current value of %q: %w", name, err)
	}

	if current == value {
		result.Changed = false
		return result, nil
	}

	path := nameToPath(name)
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("sysctl.Set: failed to write %q to %q: %w", value, name, err)
	}

	result.Changed = true
	return result, nil
}

// List walks /proc/sys recursively and returns all readable kernel parameters.
// Directories and unreadable files are silently skipped.
func List() ([]SysctlResult, error) {
	var results []SysctlResult

	walkErr := filepath.Walk(procSysRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip unreadable entries
			return nil
		}
		if info.IsDir() {
			return nil
		}

		name := pathToName(path)
		if name == "" {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			// Skip unreadable files
			return nil
		}

		results = append(results, SysctlResult{
			Name:  name,
			Value: strings.TrimSpace(string(data)),
		})
		return nil
	})

	if walkErr != nil {
		return results, fmt.Errorf("sysctl.List: walk failed: %w", walkErr)
	}
	return results, nil
}
