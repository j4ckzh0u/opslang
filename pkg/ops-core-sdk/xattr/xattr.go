//go:build !windows

// Package xattr provides extended file attribute operations.
// Supports get, set, list, and remove of extended attributes on Linux/macOS.
package xattr

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// Result represents the result of an xattr operation.
type Result struct {
	Path    string `json:"path"`
	Name    string `json:"name,omitempty"`
	Value   string `json:"value,omitempty"`
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
}

// ListResult represents the result of listing xattrs.
type ListResult struct {
	Path       string   `json:"path"`
	Attributes []string `json:"attributes"`
	Count      int      `json:"count"`
}

// GetResult represents the result of getting an xattr value.
type GetResult struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Size  int    `json:"size"`
}

// Get retrieves the value of an extended attribute.
func Get(path string, name string) (*GetResult, error) {
	if path == "" || name == "" {
		return nil, fmt.Errorf("xattr.Get: path and name are required")
	}
	result := &GetResult{Path: path, Name: name}

	// First get the size
	size, err := unix.Getxattr(path, name, nil)
	if err != nil {
		return nil, fmt.Errorf("xattr.Get: getxattr %s/%s: %w", path, name, err)
	}
	if size < 0 {
		return nil, fmt.Errorf("xattr.Get: attribute %s not found on %s", name, path)
	}

	buf := make([]byte, size)
	readSize, err := unix.Getxattr(path, name, buf)
	if err != nil {
		return nil, fmt.Errorf("xattr.Get: getxattr %s/%s: %w", path, name, err)
	}
	result.Value = string(buf[:readSize])
	result.Size = readSize
	return result, nil
}

// Set sets an extended attribute on a file.
func Set(path string, name string, value string) (*Result, error) {
	if path == "" || name == "" {
		return nil, fmt.Errorf("xattr.Set: path and name are required")
	}
	result := &Result{Path: path, Name: name, Value: value}

	// Check if value is the same
	current, err := Get(path, name)
	if err == nil && current.Value == value {
		result.Changed = false
		result.Message = "value unchanged"
		return result, nil
	}

	err = unix.Setxattr(path, name, []byte(value), 0)
	if err != nil {
		return nil, fmt.Errorf("xattr.Set: setxattr %s/%s: %w", path, name, err)
	}
	result.Changed = true
	result.Message = "attribute set"
	return result, nil
}

// Remove removes an extended attribute from a file.
func Remove(path string, name string) (*Result, error) {
	if path == "" || name == "" {
		return nil, fmt.Errorf("xattr.Remove: path and name are required")
	}
	result := &Result{Path: path, Name: name}

	// Check if attribute exists
	_, err := Get(path, name)
	if err != nil {
		result.Changed = false
		result.Message = "attribute does not exist"
		return result, nil
	}

	err = unix.Removexattr(path, name)
	if err != nil {
		return nil, fmt.Errorf("xattr.Remove: removexattr %s/%s: %w", path, name, err)
	}
	result.Changed = true
	result.Message = "attribute removed"
	return result, nil
}

// List lists all extended attributes on a file.
func List(path string) (*ListResult, error) {
	if path == "" {
		return nil, fmt.Errorf("xattr.List: path is required")
	}
	result := &ListResult{Path: path}

	// Check file exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("xattr.List: %w", err)
	}

	// Get size of attribute list
	size, err := unix.Listxattr(path, nil)
	if err != nil {
		return nil, fmt.Errorf("xattr.List: listxattr %s: %w", path, err)
	}
	if size <= 0 {
		result.Count = 0
		return result, nil
	}

	buf := make([]byte, size)
	readSize, err := unix.Listxattr(path, buf)
	if err != nil {
		return nil, fmt.Errorf("xattr.List: listxattr %s: %w", path, err)
	}

	// Attributes are null-separated
	attrs := strings.Split(string(buf[:readSize]), "\x00")
	for _, a := range attrs {
		if a != "" {
			result.Attributes = append(result.Attributes, a)
		}
	}
	result.Count = len(result.Attributes)
	return result, nil
}
