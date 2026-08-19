// Package sysfs provides sysfs attribute management operations.
package sysfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ActionResult represents the result of a sysfs operation.
type ActionResult struct {
	Path    string `json:"path"`
	Changed bool   `json:"changed"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// AttributeInfo represents information about a sysfs attribute.
type AttributeInfo struct {
	Path  string `json:"path"`
	Value string `json:"value"`
	Mode  string `json:"mode"`
}

// ListResult represents the result of listing sysfs attributes.
type ListResult struct {
	Attributes []AttributeInfo `json:"attributes"`
}

// Read reads a sysfs attribute value.
func Read(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Ensure path is under /sys
	if !strings.HasPrefix(path, "/sys") {
		return "", fmt.Errorf("path must be under /sys")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read sysfs attribute: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// Write writes a value to a sysfs attribute.
func Write(path string, value string) (ActionResult, error) {
	if path == "" {
		return ActionResult{}, fmt.Errorf("path is required")
	}

	// Ensure path is under /sys
	if !strings.HasPrefix(path, "/sys") {
		return ActionResult{Path: path, Success: false, Error: "path must be under /sys"}, fmt.Errorf("path must be under /sys")
	}

	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		return ActionResult{Path: path, Success: false, Error: err.Error()}, fmt.Errorf("failed to write sysfs attribute: %w", err)
	}

	return ActionResult{Path: path, Changed: true, Success: true}, nil
}

// Exists checks if a sysfs attribute exists.
func Exists(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("path is required")
	}

	// Ensure path is under /sys
	if !strings.HasPrefix(path, "/sys") {
		return false, fmt.Errorf("path must be under /sys")
	}

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Get retrieves a sysfs attribute with metadata.
func Get(path string) (AttributeInfo, error) {
	if path == "" {
		return AttributeInfo{}, fmt.Errorf("path is required")
	}

	// Ensure path is under /sys
	if !strings.HasPrefix(path, "/sys") {
		return AttributeInfo{}, fmt.Errorf("path must be under /sys")
	}

	info, err := os.Stat(path)
	if err != nil {
		return AttributeInfo{}, fmt.Errorf("failed to stat sysfs attribute: %w", err)
	}

	value, err := Read(path)
	if err != nil {
		return AttributeInfo{Path: path, Mode: info.Mode().String()}, err
	}

	return AttributeInfo{
		Path:  path,
		Value: value,
		Mode:  info.Mode().String(),
	}, nil
}

// List lists all attributes in a sysfs directory.
func List(dirPath string) (ListResult, error) {
	if dirPath == "" {
		return ListResult{}, fmt.Errorf("directory path is required")
	}

	// Ensure path is under /sys
	if !strings.HasPrefix(dirPath, "/sys") {
		return ListResult{}, fmt.Errorf("path must be under /sys")
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return ListResult{}, fmt.Errorf("failed to read sysfs directory: %w", err)
	}

	result := ListResult{Attributes: make([]AttributeInfo, 0)}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(dirPath, entry.Name())
		value, err := Read(fullPath)
		if err != nil {
			continue
		}
		attr := AttributeInfo{
			Path:  fullPath,
			Value: value,
		}
		if info, err := entry.Info(); err == nil {
			attr.Mode = info.Mode().String()
		}
		result.Attributes = append(result.Attributes, attr)
	}

	return result, nil
}

// SetDevicePower sets the power management state for a device.
func SetDevicePower(devicePath string, state string) (ActionResult, error) {
	if devicePath == "" || state == "" {
		return ActionResult{}, fmt.Errorf("device path and state are required")
	}

	// Ensure path is under /sys
	if !strings.HasPrefix(devicePath, "/sys") {
		return ActionResult{Path: devicePath, Success: false, Error: "path must be under /sys"}, fmt.Errorf("path must be under /sys")
	}

	powerPath := filepath.Join(devicePath, "power", "control")
	if err := os.WriteFile(powerPath, []byte(state), 0644); err != nil {
		return ActionResult{Path: powerPath, Success: false, Error: err.Error()}, fmt.Errorf("failed to set device power state: %w", err)
	}

	return ActionResult{Path: powerPath, Changed: true, Success: true}, nil
}

// GetDevicePower gets the power management state for a device.
func GetDevicePower(devicePath string) (string, error) {
	if devicePath == "" {
		return "", fmt.Errorf("device path is required")
	}

	// Ensure path is under /sys
	if !strings.HasPrefix(devicePath, "/sys") {
		return "", fmt.Errorf("path must be under /sys")
	}

	powerPath := filepath.Join(devicePath, "power", "control")
	return Read(powerPath)
}

// SetKernelParameter sets a kernel parameter via /proc/sys (sysctl-like).
func SetKernelParameter(param string, value string) (ActionResult, error) {
	if param == "" {
		return ActionResult{}, fmt.Errorf("parameter name is required")
	}

	// Convert dots to slashes: net.ipv4.ip_forward -> /proc/sys/net/ipv4/ip_forward
	paramPath := filepath.Join("/proc/sys", strings.ReplaceAll(param, ".", "/"))

	if err := os.WriteFile(paramPath, []byte(value), 0644); err != nil {
		return ActionResult{Path: paramPath, Success: false, Error: err.Error()}, fmt.Errorf("failed to set kernel parameter: %w", err)
	}

	return ActionResult{Path: paramPath, Changed: true, Success: true}, nil
}

// GetKernelParameter gets a kernel parameter value via /proc/sys.
func GetKernelParameter(param string) (string, error) {
	if param == "" {
		return "", fmt.Errorf("parameter name is required")
	}

	paramPath := filepath.Join("/proc/sys", strings.ReplaceAll(param, ".", "/"))
	return Read(paramPath)
}
