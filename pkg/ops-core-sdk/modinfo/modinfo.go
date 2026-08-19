package modinfo

import (
	"fmt"
	"os/exec"
	"strings"
)

// InfoResult represents kernel module information.
type InfoResult struct {
	Name        string            `json:"name"`
	Filename    string            `json:"filename"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
	Author      string            `json:"author,omitempty"`
	License     string            `json:"license,omitempty"`
	Depends     []string          `json:"depends,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// ListResult represents a list of loaded modules.
type ListResult struct {
	Modules []string `json:"modules"`
	Count   int      `json:"count"`
	Error   string   `json:"error,omitempty"`
}

// VersionResult represents modinfo version output.
type VersionResult struct {
	Version string `json:"version"`
	Error   string `json:"error,omitempty"`
}

// Info retrieves information about a kernel module.
func Info(module string) InfoResult {
	if module == "" {
		return InfoResult{Error: "module name is required"}
	}

	cmd := exec.Command("modinfo", module)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return InfoResult{
			Name:  module,
			Error: fmt.Sprintf("modinfo failed: %s: %s", err, string(out)),
		}
	}

	result := InfoResult{
		Name:       module,
		Parameters: make(map[string]string),
	}

	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "filename":
			result.Filename = value
		case "version":
			result.Version = value
		case "description":
			result.Description = value
		case "author":
			result.Author = value
		case "license":
			result.License = value
		case "depends":
			if value != "" {
				result.Depends = strings.Split(value, ",")
			}
		case "parm":
			// Parse parameter descriptions
			result.Parameters[key] = value
		}
	}

	return result
}

// List returns a list of currently loaded kernel modules.
func List() ListResult {
	cmd := exec.Command("lsmod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ListResult{
			Error: fmt.Sprintf("lsmod failed: %s: %s", err, string(out)),
		}
	}

	var modules []string
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			modules = append(modules, fields[0])
		}
	}

	return ListResult{
		Modules: modules,
		Count:   len(modules),
	}
}

// Version returns the modinfo version.
func Version() VersionResult {
	cmd := exec.Command("modinfo", "-V")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return VersionResult{
			Error: fmt.Sprintf("modinfo -V failed: %s: %s", err, string(out)),
		}
	}

	return VersionResult{
		Version: strings.TrimSpace(string(out)),
	}
}
