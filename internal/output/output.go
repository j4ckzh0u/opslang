// Package output provides structured output types for OpsLang execution results.
// It defines the canonical JSON schema used by opsctl deploy to report
// per-host and aggregated task outcomes.
package output

import (
	"encoding/json"
	"time"
)

// DeployResult is the top-level output of opsctl deploy.
// It captures the full execution summary across all target hosts.
type DeployResult struct {
	TaskID     string                 `json:"task_id"`
	Script     string                 `json:"script"`
	Targets    []string               `json:"targets"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
	Status     string                 `json:"status"`
	Results    map[string]*HostResult `json:"results"`
	AuditLog   string                 `json:"audit_log,omitempty"`
}

// HostResult captures per-host execution results.
type HostResult struct {
	Status   string                 `json:"status"`
	ExitCode int                    `json:"exit_code"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

// OutputEntry represents a single structured output entry emitted during
// script execution (report, alert, log, metric).
type OutputEntry struct {
	Type string      `json:"type"` // "report", "alert", "log", "metric"
	Data interface{} `json:"data"`
}

// FormatResult marshals a DeployResult to indented JSON.
func FormatResult(result *DeployResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

// MergeHostResults combines multiple host results into a DeployResult.
// The overall status is derived from per-host statuses:
//   - "no_targets" when results is empty
//   - "success" when every host succeeded
//   - "partial" when at least one host succeeded
//   - "failed" when no hosts succeeded
func MergeHostResults(taskID, script string, targets []string, results map[string]*HostResult) *DeployResult {
	now := time.Now().UTC()
	dr := &DeployResult{
		TaskID:     taskID,
		Script:     script,
		Targets:    targets,
		StartedAt:  now,
		FinishedAt: now,
		Results:    results,
	}
	dr.Status = ComputeStatus(results)
	return dr
}

// ComputeStatus derives the overall deployment status from per-host results.
func ComputeStatus(results map[string]*HostResult) string {
	if len(results) == 0 {
		return "no_targets"
	}
	allSuccess := true
	anySuccess := false
	for _, r := range results {
		if r.Status == "success" {
			anySuccess = true
		} else {
			allSuccess = false
		}
	}
	switch {
	case allSuccess:
		return "success"
	case anySuccess:
		return "partial"
	default:
		return "failed"
	}
}
