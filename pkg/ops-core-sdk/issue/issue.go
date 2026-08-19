package issue

import (
	"fmt"
	"os"
)

// ReadResult represents the result of reading /etc/issue.
type ReadResult struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// WriteResult represents the result of writing /etc/issue.
type WriteResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

const issueFile = "/etc/issue"

// Read reads the current /etc/issue content.
func Read() ReadResult {
	content, err := os.ReadFile(issueFile)
	if err != nil {
		return ReadResult{
			Error: fmt.Sprintf("failed to read %s: %s", issueFile, err),
		}
	}

	return ReadResult{
		Content: string(content),
	}
}

// Write writes new /etc/issue content.
func Write(content string) WriteResult {
	// Read current content to check if changed
	current, _ := os.ReadFile(issueFile)
	changed := string(current) != content

	err := os.WriteFile(issueFile, []byte(content), 0644)
	if err != nil {
		return WriteResult{
			Error: fmt.Sprintf("failed to write %s: %s", issueFile, err),
		}
	}

	return WriteResult{
		Changed: changed,
	}
}
