// Package patch provides patch file application operations.
// Supports applying and reversing unified diff patches.
package patch

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Result represents the result of a patch operation.
type Result struct {
	Applied  bool   `json:"applied"`
	Reversed bool   `json:"reversed"`
	File     string `json:"file,omitempty"`
	Hunks    int    `json:"hunks"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// DryRunResult represents the result of a dry-run patch check.
type DryRunResult struct {
	CanApply bool   `json:"can_apply"`
	File     string `json:"file,omitempty"`
	Hunks    int    `json:"hunks"`
	Message  string `json:"message,omitempty"`
}

// hunk represents a single diff hunk.
type hunk struct {
	oldStart int
	oldCount int
	newStart int
	newCount int
	lines    []string
}

// parsePatch parses a unified diff patch file into hunks.
func parsePatch(patchContent string) (string, []hunk, error) {
	scanner := bufio.NewScanner(strings.NewReader(patchContent))
	var fileName string
	var hunks []hunk
	var current *hunk

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "+++ "):
			// Target file
			f := strings.TrimPrefix(line, "+++ ")
			if strings.HasPrefix(f, "b/") {
				f = f[2:]
			}
			fileName = f
		case strings.HasPrefix(line, "@@"):
			if current != nil {
				hunks = append(hunks, *current)
			}
			current = &hunk{}
			fmt.Sscanf(line, "@@ -%d,%d +%d,%d", &current.oldStart, &current.oldCount, &current.newStart, &current.newCount)
		case current != nil:
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || line == "" {
				current.lines = append(current.lines, line)
			}
		}
	}
	if current != nil {
		hunks = append(hunks, *current)
	}
	if fileName == "" {
		return "", nil, fmt.Errorf("patch.parsePatch: no target file found in patch")
	}
	return fileName, hunks, nil
}

// Apply applies a patch from patch content to the filesystem.
// If reverse is true, the patch is applied in reverse.
func Apply(patchContent string, reverse bool) (*Result, error) {
	if patchContent == "" {
		return nil, fmt.Errorf("patch.Apply: patch content is empty")
	}
	result := &Result{Reversed: reverse}

	fileName, hunks, err := parsePatch(patchContent)
	if err != nil {
		return nil, err
	}
	result.File = fileName
	result.Hunks = len(hunks)

	// Read target file
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("patch.Apply: cannot read %s: %w", fileName, err)
	}
	lines := strings.Split(string(content), "\n")

	// Apply hunks (simple line-by-line replacement)
	for _, h := range hunks {
		if reverse {
			// In reverse mode, remove "+" lines and restore "-" lines
			newLines, applyErr := applyReverseHunk(lines, h)
			if applyErr != nil {
				result.Error = applyErr.Error()
				return result, applyErr
			}
			lines = newLines
		} else {
			newLines, applyErr := applyHunk(lines, h)
			if applyErr != nil {
				result.Error = applyErr.Error()
				return result, applyErr
			}
			lines = newLines
		}
	}

	// Write back
	if err := os.WriteFile(fileName, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return nil, fmt.Errorf("patch.Apply: cannot write %s: %w", fileName, err)
	}
	result.Applied = true
	result.Message = fmt.Sprintf("applied %d hunk(s) to %s", len(hunks), fileName)
	return result, nil
}

// applyHunk applies a single hunk in forward direction.
func applyHunk(lines []string, h hunk) ([]string, error) {
	start := h.oldStart - 1
	if start < 0 {
		start = 0
	}
	if start > len(lines) {
		return lines, fmt.Errorf("hunk start %d beyond file length %d", start, len(lines))
	}

	var result []string
	result = append(result, lines[:start]...)

	for _, l := range h.lines {
		switch {
		case strings.HasPrefix(l, " "):
			// Context line - keep
			if start < len(lines) {
				result = append(result, lines[start])
				start++
			}
		case strings.HasPrefix(l, "-"):
			// Remove line
			if start < len(lines) {
				start++
			}
		case strings.HasPrefix(l, "+"):
			// Add line
			result = append(result, l[1:])
		default:
			// Empty or other - treat as context
			if start < len(lines) {
				result = append(result, lines[start])
				start++
			}
		}
	}
	result = append(result, lines[start:]...)
	return result, nil
}

// applyReverseHunk applies a single hunk in reverse direction.
func applyReverseHunk(lines []string, h hunk) ([]string, error) {
	start := h.newStart - 1
	if start < 0 {
		start = 0
	}
	if start > len(lines) {
		return lines, fmt.Errorf("hunk start %d beyond file length %d", start, len(lines))
	}

	var result []string
	result = append(result, lines[:start]...)

	for _, l := range h.lines {
		switch {
		case strings.HasPrefix(l, " "):
			// Context line - keep
			if start < len(lines) {
				result = append(result, lines[start])
				start++
			}
		case strings.HasPrefix(l, "+"):
			// In reverse, "+" lines are removed
			if start < len(lines) {
				start++
			}
		case strings.HasPrefix(l, "-"):
			// In reverse, "-" lines are added back
			result = append(result, l[1:])
		default:
			if start < len(lines) {
				result = append(result, lines[start])
				start++
			}
		}
	}
	result = append(result, lines[start:]...)
	return result, nil
}

// DryRun checks if a patch can be applied without actually applying it.
func DryRun(patchContent string) (*DryRunResult, error) {
	if patchContent == "" {
		return nil, fmt.Errorf("patch.DryRun: patch content is empty")
	}
	result := &DryRunResult{}

	fileName, hunks, err := parsePatch(patchContent)
	if err != nil {
		return nil, err
	}
	result.File = fileName
	result.Hunks = len(hunks)

	// Check target file exists
	if _, err := os.Stat(fileName); err != nil {
		result.CanApply = false
		result.Message = fmt.Sprintf("target file %s does not exist", fileName)
		return result, nil
	}

	// Basic validation: check hunk counts match
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("patch.DryRun: cannot read %s: %w", fileName, err)
	}
	lines := strings.Split(string(content), "\n")

	for _, h := range hunks {
		if h.oldStart > len(lines)+1 {
			result.CanApply = false
			result.Message = fmt.Sprintf("hunk start %d beyond file length %d", h.oldStart, len(lines))
			return result, nil
		}
	}

	result.CanApply = true
	result.Message = fmt.Sprintf("patch can be applied to %s (%d hunks)", fileName, len(hunks))
	return result, nil
}
