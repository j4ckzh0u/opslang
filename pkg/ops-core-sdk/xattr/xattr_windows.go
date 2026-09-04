//go:build windows

package xattr

import "fmt"

type Result struct {
	Path    string `json:"path"`
	Name    string `json:"name,omitempty"`
	Value   string `json:"value,omitempty"`
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
}

type ListResult struct {
	Path       string   `json:"path"`
	Attributes []string `json:"attributes"`
	Count      int      `json:"count"`
}

type GetResult struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Size  int    `json:"size"`
}

func unsupported(operation string) error {
	return fmt.Errorf("xattr.%s: extended attributes are unsupported on Windows", operation)
}

func Get(path, name string) (*GetResult, error) {
	if path == "" || name == "" {
		return nil, fmt.Errorf("xattr.Get: path and name are required")
	}
	return nil, unsupported("Get")
}

func Set(path, name, value string) (*Result, error) {
	if path == "" || name == "" {
		return nil, fmt.Errorf("xattr.Set: path and name are required")
	}
	return nil, unsupported("Set")
}

func Remove(path, name string) (*Result, error) {
	if path == "" || name == "" {
		return nil, fmt.Errorf("xattr.Remove: path and name are required")
	}
	return nil, unsupported("Remove")
}

func List(path string) (*ListResult, error) {
	if path == "" {
		return nil, fmt.Errorf("xattr.List: path is required")
	}
	return nil, unsupported("List")
}
