//go:build opssec

// Package security implements permission checks and audit logging.
package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	TaskID     string                 `json:"task_id"`
	Script     string                 `json:"script"`
	Privilege  string                 `json:"privilege"`
	Targets    []string               `json:"targets"`
	User       string                 `json:"user"`
	Mode       string                 `json:"mode"`
	DryRun     bool                   `json:"dry_run"`
	Status     string                 `json:"status"`
	Results    map[string]interface{} `json:"results,omitempty"`
	Error      string                 `json:"error,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
	// Approval records the pre-execution approval decision when one was
	// required (privileged run against production targets). It makes the
	// approval event traceable in the same trail as the run itself.
	Approval *ApprovalRecord `json:"approval,omitempty"`
}

// AuditLogger handles audit log writing.
type AuditLogger struct {
	logDir string
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger(logDir string) *AuditLogger {
	if logDir == "" {
		// OPSLANG_AUDIT_DIR overrides the default location so operators
		// can keep the trail with their run artifacts.
		if d := os.Getenv("OPSLANG_AUDIT_DIR"); d != "" {
			logDir = d
		} else {
			logDir = "/var/log/opsctl"
		}
	}
	return &AuditLogger{logDir: logDir}
}

// Log writes an audit entry to the log file.
func (l *AuditLogger) Log(entry *AuditEntry) error {
	// Ensure log directory exists
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		// Fall back to temp directory if permission denied
		if os.IsPermission(err) {
			l.logDir = filepath.Join(os.TempDir(), "opsctl-logs")
			if err := os.MkdirAll(l.logDir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory: %w", err)
			}
		} else {
			// Try temp directory for any other error too
			l.logDir = filepath.Join(os.TempDir(), "opsctl-logs")
			if mkErr := os.MkdirAll(l.logDir, 0755); mkErr != nil {
				return fmt.Errorf("failed to create log directory: %w (original: %v)", mkErr, err)
			}
		}
	}

	// Create log file with date
	date := entry.Timestamp.Format("2006-01-02")
	logFile := filepath.Join(l.logDir, fmt.Sprintf("audit-%s.json", date))

	// Append entry
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	// Success - log file path for debugging
	fmt.Fprintf(os.Stderr, "Audit log written to: %s\n", logFile)
	return nil
}

// NewAuditEntry creates a new audit entry with the given parameters.
func NewAuditEntry(taskID, script, privilege string, targets []string, user, mode string, dryRun bool) *AuditEntry {
	now := time.Now().UTC()
	return &AuditEntry{
		Timestamp:  now,
		TaskID:     taskID,
		Script:     script,
		Privilege:  privilege,
		Targets:    targets,
		User:       user,
		Mode:       mode,
		DryRun:     dryRun,
		StartedAt:  now,
		FinishedAt: now,
		Results:    make(map[string]interface{}),
	}
}

// SetResult sets the result for a specific target.
func (e *AuditEntry) SetResult(target string, result interface{}) {
	e.Results[target] = result
}

// SetStatus sets the overall status and duration.
func (e *AuditEntry) SetStatus(status string, durationMs int64) {
	e.Status = status
	e.DurationMs = durationMs
	e.FinishedAt = time.Now().UTC()
}

// SetError sets the error message.
func (e *AuditEntry) SetError(err error) {
	if err != nil {
		e.Error = err.Error()
		e.Status = "failed"
	}
}
