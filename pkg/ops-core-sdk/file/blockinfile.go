// BlockInFile operations — Ansible blockinfile module equivalent.
package file

import (
	"fmt"
	"os"
	"strings"
)

// BlockInFileResult is returned by BlockInFile.
type BlockInFileResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
}

// BlockInFile inserts, updates, or removes a delimited text block in a file.
// The block is identified by begin/end markers. If present=true and the block
// exists, its content is updated. If present=false, the block is removed.
func BlockInFile(path string, marker string, content string, present bool, insertAfter string, insertBefore string) (BlockInFileResult, error) {
	if marker == "" {
		marker = "# BEGIN ANSIBLE MANAGED BLOCK"
	}
	// Parse marker: default is "# BEGIN ANSIBLE MANAGED BLOCK" / "# END ANSIBLE MANAGED BLOCK"
	// User can supply "{mark}" placeholder; we split on first newline if only one line given.
	beginMarker := marker
	endMarker := strings.Replace(marker, "BEGIN", "END", 1)

	// If marker contains {mark} placeholder (Ansible style), expand it
	if strings.Contains(marker, "{mark}") {
		beginMarker = strings.Replace(marker, "{mark}", "BEGIN", 1)
		endMarker = strings.Replace(marker, "{mark}", "END", 1)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist; create it if we need to insert
		if !os.IsNotExist(err) {
			return BlockInFileResult{}, fmt.Errorf("file.BlockInFile: %w", err)
		}
		if !present {
			return BlockInFileResult{Changed: false}, nil
		}
		data = []byte{}
	}

	lines := strings.Split(string(data), "\n")
	// Remove trailing empty element from final newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Find existing block
	blockStart := -1
	blockEnd := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == beginMarker {
			blockStart = i
		}
		if trimmed == endMarker && blockStart >= 0 {
			blockEnd = i
			break
		}
	}

	if !present {
		// Remove block
		if blockStart < 0 {
			return BlockInFileResult{Changed: false}, nil
		}
		newLines := append(lines[:blockStart], lines[blockEnd+1:]...)
		return writeLines(path, newLines, BlockInFileResult{Changed: true, Message: "block removed"})
	}

	// Present: insert or update
	blockLines := []string{beginMarker}
	for _, l := range strings.Split(content, "\n") {
		blockLines = append(blockLines, l)
	}
	blockLines = append(blockLines, endMarker)

	if blockStart >= 0 {
		// Update existing block
		newLines := make([]string, 0, len(lines)-(blockEnd-blockStart+1)+len(blockLines))
		newLines = append(newLines, lines[:blockStart]...)
		newLines = append(newLines, blockLines...)
		newLines = append(newLines, lines[blockEnd+1:]...)
		old := strings.Join(lines[blockStart:blockEnd+1], "\n")
		new := strings.Join(blockLines, "\n")
		if old == new {
			return BlockInFileResult{Changed: false}, nil
		}
		return writeLines(path, newLines, BlockInFileResult{Changed: true, Message: "block updated"})
	}

	// Insert new block
	insertIdx := -1
	if insertAfter != "" {
		for i, line := range lines {
			if strings.Contains(line, insertAfter) {
				insertIdx = i + 1
			}
		}
		if insertIdx < 0 {
			insertIdx = len(lines) // append at end
		}
	} else if insertBefore != "" {
		for i, line := range lines {
			if strings.Contains(line, insertBefore) {
				insertIdx = i
				break
			}
		}
		if insertIdx < 0 {
			insertIdx = 0 // prepend
		}
	} else {
		insertIdx = len(lines) // default: append at end
	}

	newLines := make([]string, 0, len(lines)+len(blockLines))
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, blockLines...)
	newLines = append(newLines, lines[insertIdx:]...)
	return writeLines(path, newLines, BlockInFileResult{Changed: true, Message: "block inserted"})
}

func writeLines(path string, lines []string, result BlockInFileResult) (BlockInFileResult, error) {
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return BlockInFileResult{}, fmt.Errorf("file.BlockInFile: %w", err)
	}
	return result, nil
}
