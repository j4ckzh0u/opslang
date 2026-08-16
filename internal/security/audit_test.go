package security

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewAuditLogger(t *testing.T) {
	tests := []struct {
		name   string
		logDir string
		want   string
	}{
		{"custom dir", "/tmp/test-logs", "/tmp/test-logs"},
		{"empty defaults to /var/log/opsctl", "", "/var/log/opsctl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewAuditLogger(tt.logDir)
			if l.logDir != tt.want {
				t.Errorf("logDir = %q, want %q", l.logDir, tt.want)
			}
		})
	}
}

func TestNewAuditEntry(t *testing.T) {
	entry := NewAuditEntry("task1", "script.ops", "admin", []string{"host1", "host2"}, "user1", "runner", true)

	if entry.TaskID != "task1" {
		t.Errorf("TaskID = %q, want %q", entry.TaskID, "task1")
	}
	if entry.Script != "script.ops" {
		t.Errorf("Script = %q, want %q", entry.Script, "script.ops")
	}
	if entry.Privilege != "admin" {
		t.Errorf("Privilege = %q, want %q", entry.Privilege, "admin")
	}
	if len(entry.Targets) != 2 {
		t.Errorf("Targets len = %d, want 2", len(entry.Targets))
	}
	if entry.User != "user1" {
		t.Errorf("User = %q, want %q", entry.User, "user1")
	}
	if entry.Mode != "runner" {
		t.Errorf("Mode = %q, want %q", entry.Mode, "runner")
	}
	if !entry.DryRun {
		t.Error("DryRun = false, want true")
	}
	if entry.Status != "" {
		t.Errorf("Status = %q, want empty", entry.Status)
	}
	if entry.Results == nil {
		t.Error("Results should be initialized, got nil")
	}
}

func TestAuditEntrySetResult(t *testing.T) {
	entry := NewAuditEntry("t1", "s.ops", "read_only", nil, "u", "aot", false)
	entry.SetResult("host1", map[string]string{"status": "ok"})
	entry.SetResult("host2", "failed")

	if len(entry.Results) != 2 {
		t.Fatalf("Results len = %d, want 2", len(entry.Results))
	}
	if v, ok := entry.Results["host1"].(map[string]string); !ok || v["status"] != "ok" {
		t.Errorf("host1 result unexpected: %v", entry.Results["host1"])
	}
	if v, ok := entry.Results["host2"].(string); !ok || v != "failed" {
		t.Errorf("host2 result unexpected: %v", entry.Results["host2"])
	}
}

func TestAuditEntrySetStatus(t *testing.T) {
	entry := NewAuditEntry("t1", "s.ops", "read_only", nil, "u", "aot", false)
	before := entry.FinishedAt
	time.Sleep(1 * time.Millisecond)
	entry.SetStatus("success", 500)

	if entry.Status != "success" {
		t.Errorf("Status = %q, want %q", entry.Status, "success")
	}
	if entry.DurationMs != 500 {
		t.Errorf("DurationMs = %d, want 500", entry.DurationMs)
	}
	if !entry.FinishedAt.After(before) && !entry.FinishedAt.Equal(before) {
		t.Error("FinishedAt should be updated")
	}
}

func TestAuditEntrySetError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus string
		wantErr    string
	}{
		{"non-nil error", errors.New("something failed"), "failed", "something failed"},
		{"nil error leaves status unchanged", nil, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewAuditEntry("t1", "s.ops", "read_only", nil, "u", "aot", false)
			entry.SetError(tt.err)
			if entry.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", entry.Status, tt.wantStatus)
			}
			if entry.Error != tt.wantErr {
				t.Errorf("Error = %q, want %q", entry.Error, tt.wantErr)
			}
		})
	}
}

func TestAuditLoggerLog(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "audit-logs")
	logger := NewAuditLogger(logDir)

	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	entry := &AuditEntry{
		Timestamp: now,
		TaskID:    "task-abc",
		Script:    "test.ops",
		Privilege: "admin",
		Targets:   []string{"h1"},
		User:      "tester",
		Mode:      "runner",
		DryRun:    false,
		Status:    "success",
		Results:   map[string]interface{}{"h1": "ok"},
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log() error: %v", err)
	}

	// Verify file was created with correct name
	logFile := filepath.Join(logDir, "audit-2026-08-16.json")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Parse JSON lines
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var parsed AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed.TaskID != "task-abc" {
		t.Errorf("parsed TaskID = %q, want %q", parsed.TaskID, "task-abc")
	}

	// Append another entry
	entry2 := &AuditEntry{
		Timestamp: now,
		TaskID:    "task-def",
		Script:    "test2.ops",
		Privilege: "read_only",
		Status:    "failed",
	}
	if err := logger.Log(entry2); err != nil {
		t.Fatalf("second Log() error: %v", err)
	}

	data, _ = os.ReadFile(logFile)
	lines = strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestAuditLoggerLogPermissionDeniedFallsBack(t *testing.T) {
	// Use a path that will fail to create (under a file, not a dir)
	tmpDir := t.TempDir()
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	// logDir is a file, so MkdirAll will fail
	logger := NewAuditLogger(blocker)

	entry := &AuditEntry{
		Timestamp: time.Now(),
		TaskID:    "t1",
		Status:    "ok",
	}

	// Should succeed by falling back to temp dir
	err := logger.Log(entry)
	if err != nil {
		t.Fatalf("expected fallback to succeed, got: %v", err)
	}

	// Verify logger's logDir was changed to fallback
	if !strings.Contains(logger.logDir, "opsctl-logs") {
		t.Errorf("logDir = %q, expected fallback to opsctl-logs", logger.logDir)
	}
}

func TestAuditEntryJSON(t *testing.T) {
	entry := &AuditEntry{
		Timestamp:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TaskID:     "t1",
		Script:     "s.ops",
		Privilege:  "root",
		Targets:    []string{"h1"},
		User:       "u",
		Mode:       "aot",
		DryRun:     true,
		Status:     "success",
		DurationMs: 100,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed AuditEntry
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if parsed.TaskID != "t1" || parsed.DurationMs != 100 || !parsed.DryRun {
		t.Errorf("roundtrip mismatch: %+v", parsed)
	}
}
