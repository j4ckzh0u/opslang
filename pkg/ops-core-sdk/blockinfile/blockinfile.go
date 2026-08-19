// Package blockinfile manages blocks of text in files.
// Equivalent to ansible.builtin.blockinfile module.
package blockinfile

import (
	"fmt"
	"os"
	"strings"
)

// Result is returned by all blockinfile functions.
type Result struct {
	Status  string `json:"status"`            // success/failed
	Changed bool   `json:"changed"`           // whether the file was modified
	Path    string `json:"path"`              // file path
	Msg     string `json:"msg,omitempty"`     // operation description
	Error   string `json:"error,omitempty"`
}

const defaultMarker = "# {mark} ANSIBLE MANAGED BLOCK"

func markerLine(marker, mark string) string {
	return strings.Replace(marker, "{mark}", mark, 1)
}

// Manage inserts, updates, or removes a block of text delimited by markers.
// state: "present" or "absent".
// block: the text content to insert (without markers).
// marker: custom marker template with {mark} placeholder.
// insertAfter/insertBefore: regex patterns for insertion point (simplified: line substring match).
func Manage(path, block, state, marker, insertAfter, insertBefore string) Result {
	if path == "" {
		return Result{Status: "failed", Error: "path is required"}
	}
	if marker == "" {
		marker = defaultMarker
	}

	beginMarker := markerLine(marker, "BEGIN")
	endMarker := markerLine(marker, "END")

	content, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return Result{Status: "failed", Path: path, Error: fmt.Sprintf("read file: %v", err)}
		}
		content = []byte{}
	}

	lines := strings.Split(string(content), "\n")
	// Remove trailing empty line from split
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Find existing block
	beginIdx, endIdx := -1, -1
	for i, line := range lines {
		if strings.TrimSpace(line) == strings.TrimSpace(beginMarker) {
			beginIdx = i
		}
		if strings.TrimSpace(line) == strings.TrimSpace(endMarker) {
			endIdx = i
		}
	}

	switch state {
	case "absent":
		if beginIdx == -1 || endIdx == -1 {
			return Result{Status: "success", Changed: false, Path: path, Msg: "block not found, nothing to remove"}
		}
		// Remove block including markers
		newLines := append(lines[:beginIdx], lines[endIdx+1:]...)
		if err := writeLines(path, newLines); err != nil {
			return Result{Status: "failed", Path: path, Error: fmt.Sprintf("write file: %v", err)}
		}
		return Result{Status: "success", Changed: true, Path: path, Msg: "block removed"}

	case "present":
		blockLines := strings.Split(block, "\n")
		newBlock := make([]string, 0, len(blockLines)+2)
		newBlock = append(newBlock, beginMarker)
		newBlock = append(newBlock, blockLines...)
		newBlock = append(newBlock, endMarker)

		if beginIdx != -1 && endIdx != -1 {
			// Check if existing block content matches
			existingBlock := lines[beginIdx+1 : endIdx]
			if linesEqual(existingBlock, blockLines) {
				return Result{Status: "success", Changed: false, Path: path, Msg: "block already up to date"}
			}
			// Replace existing block
			newLines := make([]string, 0, len(lines)-(endIdx-beginIdx+1)+len(newBlock))
			newLines = append(newLines, lines[:beginIdx]...)
			newLines = append(newLines, newBlock...)
			newLines = append(newLines, lines[endIdx+1:]...)
			if err := writeLines(path, newLines); err != nil {
				return Result{Status: "failed", Path: path, Error: fmt.Sprintf("write file: %v", err)}
			}
			return Result{Status: "success", Changed: true, Path: path, Msg: "block updated"}
		}

		// Insert new block
		insertIdx := len(lines) // default: append at end
		if insertAfter != "" {
			for i, line := range lines {
				if strings.Contains(line, insertAfter) {
					insertIdx = i + 1
				}
			}
		} else if insertBefore != "" {
			for i, line := range lines {
				if strings.Contains(line, insertBefore) {
					insertIdx = i
					break
				}
			}
		}

		newLines := make([]string, 0, len(lines)+len(newBlock))
		newLines = append(newLines, lines[:insertIdx]...)
		newLines = append(newLines, newBlock...)
		newLines = append(newLines, lines[insertIdx:]...)
		if err := writeLines(path, newLines); err != nil {
			return Result{Status: "failed", Path: path, Error: fmt.Sprintf("write file: %v", err)}
		}
		return Result{Status: "success", Changed: true, Path: path, Msg: "block inserted"}

	default:
		return Result{Status: "failed", Path: path, Error: "state must be 'present' or 'absent'"}
	}
}

func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}

func linesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Read reads the current block content between markers.
func Read(path, marker string) (string, bool, error) {
	if marker == "" {
		marker = defaultMarker
	}
	beginMarker := markerLine(marker, "BEGIN")
	endMarker := markerLine(marker, "END")

	content, err := os.ReadFile(path)
	if err != nil {
		return "", false, err
	}

	lines := strings.Split(string(content), "\n")
	beginIdx, endIdx := -1, -1
	for i, line := range lines {
		if strings.TrimSpace(line) == strings.TrimSpace(beginMarker) {
			beginIdx = i
		}
		if strings.TrimSpace(line) == strings.TrimSpace(endMarker) {
			endIdx = i
		}
	}

	if beginIdx == -1 || endIdx == -1 || endIdx <= beginIdx {
		return "", false, nil
	}

	block := strings.Join(lines[beginIdx+1:endIdx], "\n")
	return block, true, nil
}
