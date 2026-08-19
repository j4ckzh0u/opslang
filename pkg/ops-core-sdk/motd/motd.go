package motd

import (
	"fmt"
	"os"
)

// ReadResult represents the result of reading MOTD.
type ReadResult struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// WriteResult represents the result of writing MOTD.
type WriteResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

const motdFile = "/etc/motd"

// Read reads the current MOTD content.
func Read() ReadResult {
	content, err := os.ReadFile(motdFile)
	if err != nil {
		return ReadResult{
			Error: fmt.Sprintf("failed to read %s: %s", motdFile, err),
		}
	}

	return ReadResult{
		Content: string(content),
	}
}

// Write writes new MOTD content.
func Write(content string) WriteResult {
	// Read current content to check if changed
	current, _ := os.ReadFile(motdFile)
	changed := string(current) != content

	err := os.WriteFile(motdFile, []byte(content), 0644)
	if err != nil {
		return WriteResult{
			Error: fmt.Sprintf("failed to write %s: %s", motdFile, err),
		}
	}

	return WriteResult{
		Changed: changed,
	}
}
