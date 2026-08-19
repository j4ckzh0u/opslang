// Package lineinfile provides line-level file editing operations.
package lineinfile

import (
	"bufio"
	"fmt"
	"os"
	goregexp "regexp"
	"strings"
)

// Result represents the result of a lineinfile operation.
type Result struct {
	Changed bool   `json:"changed"`
	Path    string `json:"path"`
	Line    string `json:"line,omitempty"`
	Message string `json:"message"`
}

// Ensure ensures a line exists in a file, matching by regexp if provided.
func Ensure(path string, line string, regexp string, create bool) (*Result, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		if !create {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		content = []byte{}
	}

	lines := strings.Split(string(content), "\n")

	// Check if line already exists
	for _, l := range lines {
		if l == line {
			return &Result{Changed: false, Path: path, Line: line, Message: "Line already exists"}, nil
		}
	}

	// If regexp provided, replace matching line
	if regexp != "" {
		re, err := goregexp.Compile(regexp)
		if err != nil {
			return nil, fmt.Errorf("invalid regexp: %w", err)
		}
		for i, l := range lines {
			if re.MatchString(l) {
				lines[i] = line
				newContent := strings.Join(lines, "\n")
				if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
					return nil, fmt.Errorf("failed to write file: %w", err)
				}
				return &Result{Changed: true, Path: path, Line: line, Message: "Line replaced"}, nil
			}
		}
	}

	// Append line
	lines = append(lines, line)
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return &Result{Changed: true, Path: path, Line: line, Message: "Line added"}, nil
}

// Absent ensures a line matching the regexp is absent from the file.
func Absent(path string, regexp string) (*Result, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	re, err := goregexp.Compile(regexp)
	if err != nil {
		return nil, fmt.Errorf("invalid regexp: %w", err)
	}

	var newLines []string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	removed := false
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			removed = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !removed {
		return &Result{Changed: false, Path: path, Message: "No matching line found"}, nil
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return &Result{Changed: true, Path: path, Message: "Line removed"}, nil
}
