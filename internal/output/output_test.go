package output

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFormatResult_ProducesValidJSON(t *testing.T) {
	start := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)
	finish := time.Date(2026, 8, 15, 10, 0, 12, 0, time.UTC)

	result := &DeployResult{
		TaskID:     "abc123",
		Script:     "check_cpu.ops",
		Targets:    []string{"host1", "host2"},
		StartedAt:  start,
		FinishedAt: finish,
		Status:     "success",
		Results: map[string]*HostResult{
			"host1": {
				Status:   "success",
				ExitCode: 0,
				Data:     map[string]interface{}{"cpu": 12.5},
			},
			"host2": {
				Status:   "failed",
				ExitCode: 1,
				Error:    "timeout",
			},
		},
		AuditLog: "/var/log/opsctl/abc123.json",
	}

	data, err := FormatResult(result)
	if err != nil {
		t.Fatalf("FormatResult returned error: %v", err)
	}

	// Verify it parses back as valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("FormatResult produced invalid JSON: %v", err)
	}

	// Verify key fields survived the round-trip.
	if parsed["task_id"] != "abc123" {
		t.Errorf("task_id = %v, want abc123", parsed["task_id"])
	}
	if parsed["script"] != "check_cpu.ops" {
		t.Errorf("script = %v, want check_cpu.ops", parsed["script"])
	}
	if parsed["status"] != "success" {
		t.Errorf("status = %v, want success", parsed["status"])
	}
	if parsed["audit_log"] != "/var/log/opsctl/abc123.json" {
		t.Errorf("audit_log = %v, want /var/log/opsctl/abc123.json", parsed["audit_log"])
	}

	results, ok := parsed["results"].(map[string]interface{})
	if !ok {
		t.Fatal("results is not a map")
	}
	host1, ok := results["host1"].(map[string]interface{})
	if !ok {
		t.Fatal("host1 is not a map")
	}
	if host1["status"] != "success" {
		t.Errorf("host1 status = %v, want success", host1["status"])
	}
}

func TestFormatResult_OmitsAuditLog(t *testing.T) {
	result := &DeployResult{
		TaskID:  "t1",
		Script:  "s.ops",
		Status:  "success",
		Results: map[string]*HostResult{},
	}

	data, err := FormatResult(result)
	if err != nil {
		t.Fatalf("FormatResult returned error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, exists := parsed["audit_log"]; exists {
		t.Error("audit_log should be omitted when empty")
	}
}

func TestMergeHostResults_AllSuccess(t *testing.T) {
	results := map[string]*HostResult{
		"h1": {Status: "success", ExitCode: 0},
		"h2": {Status: "success", ExitCode: 0},
	}

	dr := MergeHostResults("task1", "script.ops", []string{"h1", "h2"}, results)

	if dr.Status != "success" {
		t.Errorf("status = %q, want success", dr.Status)
	}
	if dr.TaskID != "task1" {
		t.Errorf("taskID = %q, want task1", dr.TaskID)
	}
	if len(dr.Targets) != 2 {
		t.Errorf("targets length = %d, want 2", len(dr.Targets))
	}
}

func TestMergeHostResults_PartialSuccess(t *testing.T) {
	results := map[string]*HostResult{
		"h1": {Status: "success", ExitCode: 0},
		"h2": {Status: "failed", ExitCode: 1, Error: "timeout"},
	}

	dr := MergeHostResults("task2", "script.ops", []string{"h1", "h2"}, results)

	if dr.Status != "partial" {
		t.Errorf("status = %q, want partial", dr.Status)
	}
}

func TestMergeHostResults_AllFailed(t *testing.T) {
	results := map[string]*HostResult{
		"h1": {Status: "failed", ExitCode: 1},
		"h2": {Status: "failed", ExitCode: 2},
	}

	dr := MergeHostResults("task3", "script.ops", []string{"h1", "h2"}, results)

	if dr.Status != "failed" {
		t.Errorf("status = %q, want failed", dr.Status)
	}
}

func TestMergeHostResults_NoTargets(t *testing.T) {
	dr := MergeHostResults("task4", "script.ops", []string{}, map[string]*HostResult{})

	if dr.Status != "no_targets" {
		t.Errorf("status = %q, want no_targets", dr.Status)
	}
}

func TestComputeStatus(t *testing.T) {
	tests := []struct {
		name    string
		results map[string]*HostResult
		want    string
	}{
		{
			name:    "empty",
			results: map[string]*HostResult{},
			want:    "no_targets",
		},
		{
			name: "all success",
			results: map[string]*HostResult{
				"a": {Status: "success"},
				"b": {Status: "success"},
			},
			want: "success",
		},
		{
			name: "partial",
			results: map[string]*HostResult{
				"a": {Status: "success"},
				"b": {Status: "failed"},
			},
			want: "partial",
		},
		{
			name: "all failed",
			results: map[string]*HostResult{
				"a": {Status: "failed"},
				"b": {Status: "failed"},
			},
			want: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeStatus(tt.results)
			if got != tt.want {
				t.Errorf("ComputeStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
