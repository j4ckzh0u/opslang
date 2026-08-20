package async_status

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPollEmptyJobID(t *testing.T) {
	r := Poll("", "")
	if r.Error == "" {
		t.Error("expected error for empty job_id")
	}
}

func TestPollNonExistent(t *testing.T) {
	r := Poll("nonexistent-job", "/tmp/nonexistent-dir")
	if r.Started || r.Finished {
		t.Error("expected not started for non-existent job")
	}
}

func TestPollFinishedJob(t *testing.T) {
	dir := t.TempDir()
	jobID := "test-job-123"
	result := map[string]interface{}{
		"stdout":    "done",
		"exit_code": float64(0),
		"state":     "finished",
	}
	data, _ := json.Marshal(result)
	os.WriteFile(filepath.Join(dir, jobID), data, 0644)

	r := Poll(jobID, dir)
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Finished {
		t.Error("expected finished")
	}
	if r.State != "finished" {
		t.Errorf("unexpected state: %s", r.State)
	}
	if r.Stdout != "done" {
		t.Errorf("unexpected stdout: %q", r.Stdout)
	}
}

func TestCleanup(t *testing.T) {
	dir := t.TempDir()
	jobID := "test-cleanup"
	os.WriteFile(filepath.Join(dir, jobID), []byte("{}"), 0644)
	Cleanup(jobID, dir)
	if _, err := os.Stat(filepath.Join(dir, jobID)); !os.IsNotExist(err) {
		t.Error("expected file removed")
	}
}
