package svn

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestResultJSON tests JSON marshaling of Result.
func TestResultJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  Result
		fields []string
	}{
		{
			name:   "success",
			input:  Result{Status: "success", Changed: true, Path: "/tmp/wc", Revision: "123"},
			fields: []string{"status", "changed", "path", "revision"},
		},
		{
			name:   "failed",
			input:  Result{Status: "failed", Error: "svn not found"},
			fields: []string{"status", "error"},
		},
		{
			name:   "already up to date",
			input:  Result{Status: "success", Changed: false, Path: "/tmp/wc"},
			fields: []string{"status", "changed", "path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			var result Result
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if result.Status != tt.input.Status {
				t.Errorf("Status: got %q, want %q", result.Status, tt.input.Status)
			}
			if result.Changed != tt.input.Changed {
				t.Errorf("Changed: got %v, want %v", result.Changed, tt.input.Changed)
			}
			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("unmarshal raw failed: %v", err)
			}
			for _, f := range tt.fields {
				if _, ok := raw[f]; !ok {
					t.Errorf("missing expected field: %s", f)
				}
			}
		})
	}
}

// TestStatusEntryJSON tests JSON marshaling of StatusEntry.
func TestStatusEntryJSON(t *testing.T) {
	entry := StatusEntry{Status: "M", Path: "file.txt"}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result StatusEntry
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result.Status != "M" || result.Path != "file.txt" {
		t.Errorf("unexpected result: %+v", result)
	}
}

// TestInfoResultJSON tests JSON marshaling of InfoResult.
func TestInfoResultJSON(t *testing.T) {
	info := InfoResult{
		Status: "success",
		Path:   "/tmp/wc",
		URL:    "svn://example.com/repo/trunk",
		Repo:   "svn://example.com/repo",
		Rev:    "456",
		Node:   "directory",
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result InfoResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result.URL != info.URL {
		t.Errorf("URL: got %q, want %q", result.URL, info.URL)
	}
	if result.Rev != info.Rev {
		t.Errorf("Rev: got %q, want %q", result.Rev, info.Rev)
	}
}

// TestCheckoutValidation tests that Checkout validates required arguments.
func TestCheckoutValidation(t *testing.T) {
	_, err := Checkout("", "/tmp/dest", "", false)
	if err == nil {
		t.Error("expected error for empty url")
	}
	_, err = Checkout("svn://example.com/repo", "", "", false)
	if err == nil {
		t.Error("expected error for empty dest")
	}
}

// TestUpdateValidation tests that Update validates required arguments.
func TestUpdateValidation(t *testing.T) {
	_, err := Update("", "")
	if err == nil {
		t.Error("expected error for empty dest")
	}
}

// TestExportValidation tests that Export validates required arguments.
func TestExportValidation(t *testing.T) {
	_, err := Export("", "/tmp/dest", "", false)
	if err == nil {
		t.Error("expected error for empty url")
	}
	_, err = Export("svn://example.com/repo", "", "", false)
	if err == nil {
		t.Error("expected error for empty dest")
	}
}

// TestStatusValidation tests that Status validates required arguments.
func TestStatusValidation(t *testing.T) {
	_, err := Status("")
	if err == nil {
		t.Error("expected error for empty dest")
	}
}

// TestInfoValidation tests that Info validates required arguments.
func TestInfoValidation(t *testing.T) {
	_, err := Info("")
	if err == nil {
		t.Error("expected error for empty dest")
	}
}

// TestCleanupValidation tests that Cleanup validates required arguments.
func TestCleanupValidation(t *testing.T) {
	_, err := Cleanup("")
	if err == nil {
		t.Error("expected error for empty dest")
	}
}

// TestRevertValidation tests that Revert validates required arguments.
func TestRevertValidation(t *testing.T) {
	_, err := Revert("", false)
	if err == nil {
		t.Error("expected error for empty dest")
	}
}

// TestToolNotFound tests the toolNotFound helper returns a non-nil error.
func TestToolNotFound(t *testing.T) {
	err := toolNotFound(fmt.Errorf("exec not found"))
	if err == nil {
		t.Error("expected non-nil error")
	}
	if !strings.Contains(err.Error(), "svn command not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}
