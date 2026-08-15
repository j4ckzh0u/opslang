package security

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	TaskID     string                 `json:"task_id"`
	Script     string                 `json:"script"`
	Targets    []string               `json:"targets"`
	User       string                 `json:"user"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
	Status     string                 `json:"status"`
	Result     map[string]interface{} `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// AuditLogger writes audit entries to a JSON lines file
type AuditLogger struct {
	path string
	mu   sync.Mutex
}

// NewAuditLogger creates a new audit logger that writes to the specified path
func NewAuditLogger(path string) (*AuditLogger, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %w", err)
	}

	// Create or open the file to verify we can write to it
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	f.Close()

	return &AuditLogger{path: path}, nil
}

// DefaultAuditLogger creates an audit logger at the default location
func DefaultAuditLogger() (*AuditLogger, error) {
	// Use /var/log/opsctl/ if root, otherwise ~/.opsctl/logs/
	var logPath string
	if os.Geteuid() == 0 {
		logPath = "/var/log/opsctl/audit.jsonl"
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		logPath = filepath.Join(home, ".opsctl", "logs", "audit.jsonl")
	}

	return NewAuditLogger(logPath)
}

// Log writes an audit entry to the log file
func (l *AuditLogger) Log(entry AuditEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	return nil
}

// Load reads all audit entries from the log file
func Load(path string) ([]AuditEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer f.Close()

	var entries []AuditEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal audit entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading audit log: %w", err)
	}

	return entries, nil
}

// Path returns the path to the audit log file
func (l *AuditLogger) Path() string {
	return l.path
}
