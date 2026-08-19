// File replace operations — Ansible replace module equivalent.
package file

import (
	"fmt"
	"os"
	"regexp"
)

// ReplaceResult is returned by Replace, reporting how many replacements were made.
type ReplaceResult struct {
	Changed  bool `json:"changed"`
	Replacements int `json:"replacements"`
}

// Replace replaces all occurrences of a regex pattern in a file.
// If after is provided, only replacements within the after/before markers are applied.
func Replace(path string, pattern string, replacement string, after string, before string) (ReplaceResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ReplaceResult{}, fmt.Errorf("file.Replace: %w", err)
	}

	content := string(data)
	rx, err := regexp.Compile(pattern)
	if err != nil {
		return ReplaceResult{}, fmt.Errorf("file.Replace: invalid pattern: %w", err)
	}

	var result ReplaceResult

	// If after/before are specified, only replace within that range
	if after != "" || before != "" {
		startIdx := 0
		endIdx := len(content)

		if after != "" {
			afterRx, err := regexp.Compile(after)
			if err != nil {
				return ReplaceResult{}, fmt.Errorf("file.Replace: invalid after pattern: %w", err)
			}
			loc := afterRx.FindStringIndex(content)
			if loc == nil {
				return result, nil // no match, no change
			}
			startIdx = loc[1]
		}

		if before != "" {
			beforeRx, err := regexp.Compile(before)
			if err != nil {
				return ReplaceResult{}, fmt.Errorf("file.Replace: invalid before pattern: %w", err)
			}
			searchFrom := startIdx
			if searchFrom >= len(content) {
				return result, nil
			}
			loc := beforeRx.FindStringIndex(content[searchFrom:])
			if loc == nil {
				return result, nil
			}
			endIdx = searchFrom + loc[0]
		}

		prefix := content[:startIdx]
		middle := content[startIdx:endIdx]
		suffix := content[endIdx:]

		count := 0
		newMiddle := rx.ReplaceAllStringFunc(middle, func(match string) string {
			count++
			return replacement
		})
		result.Replacements = count
		result.Changed = count > 0

		if result.Changed {
			if err := os.WriteFile(path, []byte(prefix+newMiddle+suffix), 0644); err != nil {
				return ReplaceResult{}, fmt.Errorf("file.Replace: %w", err)
			}
		}
		return result, nil
	}

	// Simple global replace
	count := 0
	newContent := rx.ReplaceAllStringFunc(content, func(match string) string {
		count++
		return replacement
	})
	result.Replacements = count
	result.Changed = count > 0

	if result.Changed {
		if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
			return ReplaceResult{}, fmt.Errorf("file.Replace: %w", err)
		}
	}

	return result, nil
}
