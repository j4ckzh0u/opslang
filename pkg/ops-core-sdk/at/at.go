// Package at provides at job scheduling operations.
package at

import (
	"fmt"
	"os/exec"
	"strings"
)

// Job represents an at job.
type Job struct {
	Rank string `json:"rank"`
	Date string `json:"date"`
	Time string `json:"time"`
	User string `json:"user"`
	ID   string `json:"id"`
}

// ListResult represents the result of listing jobs.
type ListResult struct {
	Jobs []Job `json:"jobs"`
}

// ActionResult represents the result of an at action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
	JobID   string `json:"job_id,omitempty"`
}

// List returns all at jobs.
func List() (*ListResult, error) {
	out, err := exec.Command("atq").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("atq failed: %w", err)
	}

	result := &ListResult{Jobs: make([]Job, 0)}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			job := Job{
				Rank: fields[0],
				Date: fields[1] + " " + fields[2] + " " + fields[3],
				Time: fields[4],
				User: fields[5],
			}
			if len(fields) >= 7 {
				job.ID = fields[6]
			}
			result.Jobs = append(result.Jobs, job)
		}
	}

	return result, nil
}

// Schedule schedules a command to run at a specific time.
func Schedule(command string, timeSpec string) (*ActionResult, error) {
	cmd := exec.Command("at", timeSpec)
	cmd.Stdin = strings.NewReader(command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("at schedule failed: %w (output: %s)", err, string(out))
	}

	// Extract job ID from output
	jobID := ""
	if strings.Contains(string(out), "job") {
		parts := strings.Fields(string(out))
		for i, p := range parts {
			if p == "job" && i+1 < len(parts) {
				jobID = parts[i+1]
				break
			}
		}
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Scheduled job at %s", timeSpec),
		JobID:   jobID,
	}, nil
}

// Remove removes an at job.
func Remove(jobID string) (*ActionResult, error) {
	cmd := exec.Command("atrm", jobID)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("atrm failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Removed job %s", jobID),
	}, nil
}
