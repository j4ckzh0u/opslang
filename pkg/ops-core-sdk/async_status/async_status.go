// Package async_status provides Ansible async_status module equivalent.
// Polls for the status of an async job.
package async_status

import (
	"encoding/json"
	"os"
	"time"
)

// AsyncResult is returned by Poll.
type AsyncResult struct {
	Started    bool   `json:"started"`
	Finished   bool   `json:"finished"`
	AnsibleJobID string `json:"ansible_job_id"`
	ResultsFile  string `json:"results_file"`
	State      string `json:"state,omitempty"` // started/finished/timeout
	Stdout     string `json:"stdout,omitempty"`
	ExitCode   int    `json:"exit_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Poll checks the status of an async job by reading its results file.
func Poll(jobID string, resultsDir string) AsyncResult {
	if jobID == "" {
		return AsyncResult{Error: "job_id is required"}
	}
	if resultsDir == "" {
		resultsDir = os.TempDir()
	}
	resultsFile := resultsDir + "/" + jobID
	data, err := os.ReadFile(resultsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return AsyncResult{Started: false, Finished: false, AnsibleJobID: jobID, ResultsFile: resultsFile}
		}
		return AsyncResult{Error: err.Error(), AnsibleJobID: jobID}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return AsyncResult{Error: "invalid result file: " + err.Error(), AnsibleJobID: jobID, ResultsFile: resultsFile}
	}

	ar := AsyncResult{
		Started:      true,
		Finished:     true,
		AnsibleJobID: jobID,
		ResultsFile:  resultsFile,
		State:        "finished",
	}
	if stdout, ok := result["stdout"].(string); ok {
		ar.Stdout = stdout
	}
	if ec, ok := result["exit_code"].(float64); ok {
		ar.ExitCode = int(ec)
	}
	if s, ok := result["state"].(string); ok {
		ar.State = s
		if s == "started" {
			ar.Finished = false
		}
	}
	return ar
}

// Cleanup removes the results file for a completed job.
func Cleanup(jobID string, resultsDir string) bool {
	if resultsDir == "" {
		resultsDir = os.TempDir()
	}
	os.Remove(resultsDir + "/" + jobID)
	return true
}

// WaitFor polls until finished or timeout.
func WaitFor(jobID string, resultsDir string, timeout time.Duration, interval time.Duration) AsyncResult {
	deadline := time.Now().Add(timeout)
	if interval <= 0 {
		interval = time.Second
	}
	for time.Now().Before(deadline) {
		r := Poll(jobID, resultsDir)
		if r.Error != "" {
			return r
		}
		if r.Finished {
			return r
		}
		time.Sleep(interval)
	}
	return AsyncResult{Started: true, Finished: false, State: "timeout", AnsibleJobID: jobID, Error: "timed out waiting for async job"}
}
