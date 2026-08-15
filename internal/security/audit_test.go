package security

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAuditLoggerLogAndLoad(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	// Create logger
	logger, err := NewAuditLogger(logPath)
	if err != nil {
		t.Fatalf("NewAuditLogger failed: %v", err)
	}

	// Log an entry
	now := time.Now()
	entry := AuditEntry{
		TaskID:     "task-123",
		Script:     "test.ops",
		Targets:    []string{"host1", "host2"},
		User:       "testuser",
		StartedAt:  now,
		FinishedAt: now.Add(5 * time.Second),
		Status:     "success",
		Result:     map[string]interface{}{"key": "value"},
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Load and verify
	entries, err := Load(logPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	got := entries[0]
	if got.TaskID != entry.TaskID {
		t.Errorf("TaskID = %v, want %v", got.TaskID, entry.TaskID)
	}
	if got.Script != entry.Script {
		t.Errorf("Script = %v, want %v", got.Script, entry.Script)
	}
	if got.User != entry.User {
		t.Errorf("User = %v, want %v", got.User, entry.User)
	}
	if got.Status != entry.Status {
		t.Errorf("Status = %v, want %v", got.Status, entry.Status)
	}
	if len(got.Targets) != 2 {
		t.Errorf("Targets length = %d, want 2", len(got.Targets))
	}
}

func TestAuditLoggerMultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	logger, err := NewAuditLogger(logPath)
	if err != nil {
		t.Fatalf("NewAuditLogger failed: %v", err)
	}

	// Log multiple entries
	for i := 0; i < 5; i++ {
		entry := AuditEntry{
			TaskID: "task-" + string(rune('0'+i)),
			Script: "test.ops",
			Status: "success",
		}
		if err := logger.Log(entry); err != nil {
			t.Fatalf("Log failed at entry %d: %v", i, err)
		}
	}

	// Load and verify
	entries, err := Load(logPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(entries) != 5 {
		t.Fatalf("Expected 5 entries, got %d", len(entries))
	}
}

func TestAuditLoggerConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	logger, err := NewAuditLogger(logPath)
	if err != nil {
		t.Fatalf("NewAuditLogger failed: %v", err)
	}

	// Log from multiple goroutines concurrently
	var wg sync.WaitGroup
	numGoroutines := 10
	numLogsPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for j := 0; j < numLogsPerGoroutine; j++ {
				entry := AuditEntry{
					TaskID: "concurrent-task",
					Status: "success",
				}
				if err := logger.Log(entry); err != nil {
					t.Errorf("Log failed: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Load and verify
	entries, err := Load(logPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expectedTotal := numGoroutines * numLogsPerGoroutine
	if len(entries) != expectedTotal {
		t.Errorf("Expected %d entries, got %d", expectedTotal, len(entries))
	}
}

func TestAuditLoggerCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "nested", "dir", "audit.jsonl")

	logger, err := NewAuditLogger(logPath)
	if err != nil {
		t.Fatalf("NewAuditLogger failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(logPath)); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}

	// Verify we can log
	entry := AuditEntry{TaskID: "test", Status: "success"}
	if err := logger.Log(entry); err != nil {
		t.Errorf("Log failed: %v", err)
	}
}

func TestAuditLoggerPath(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	logger, err := NewAuditLogger(logPath)
	if err != nil {
		t.Fatalf("NewAuditLogger failed: %v", err)
	}

	if got := logger.Path(); got != logPath {
		t.Errorf("Path() = %v, want %v", got, logPath)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	// Create empty file
	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	f.Close()

	entries, err := Load(logPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/audit.jsonl")
	if err == nil {
		t.Error("Expected error loading non-existent file")
	}
}

func TestAuditEntryWithError(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	logger, err := NewAuditLogger(logPath)
	if err != nil {
		t.Fatalf("NewAuditLogger failed: %v", err)
	}

	entry := AuditEntry{
		TaskID: "failed-task",
		Status: "failed",
		Error:  "connection timeout",
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	entries, err := Load(logPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Error != "connection timeout" {
		t.Errorf("Error = %v, want %v", entries[0].Error, "connection timeout")
	}
}
