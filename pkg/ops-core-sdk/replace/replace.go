// Package replace provides file content replacement operations.
package replace

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Result represents the result of a replace operation.
type Result struct {
	Changed  bool   `json:"changed"`
	Path     string `json:"path"`
	Replaced int    `json:"replaced"`
	Message  string `json:"message"`
}

// Replace replaces all occurrences of a pattern in a file.
func Replace(path string, pattern string, replacement string, regexpMode bool) (*Result, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var newContent string
	var replaced int

	if regexpMode {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regexp: %w", err)
		}
		newContent = re.ReplaceAllString(string(content), replacement)
		replaced = len(re.FindAllString(string(content), -1))
	} else {
		newContent = strings.ReplaceAll(string(content), pattern, replacement)
		replaced = strings.Count(string(content), pattern)
	}

	if replaced == 0 {
		return &Result{Changed: false, Path: path, Replaced: 0, Message: "No matches found"}, nil
	}

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return &Result{Changed: true, Path: path, Replaced: replaced, Message: fmt.Sprintf("Replaced %d occurrences", replaced)}, nil
}
