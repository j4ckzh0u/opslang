package nomad

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// JobInfo represents Nomad job information
type JobInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	SubmitTime  int64  `json:"submit_time"`
	Namespace   string `json:"namespace"`
	Datacenters string `json:"datacenters"`
}

// AllocInfo represents Nomad allocation information
type AllocInfo struct {
	ID        string `json:"id"`
	JobID     string `json:"job_id"`
	TaskGroup string `json:"task_group"`
	NodeID    string `json:"node_id"`
	Status    string `json:"status"`
	Desired   string `json:"desired"`
	Client    string `json:"client"`
}

// NodeInfo represents Nomad node information
type NodeInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Datacenter string `json:"datacenter"`
	Status     string `json:"status"`
	Drain      bool   `json:"drain"`
	Eligible   bool   `json:"eligible"`
}

// NomadResult represents Nomad operation result
type NomadResult struct {
	Changed    bool        `json:"changed"`
	ID         string      `json:"id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
	DurationMs int64       `json:"duration_ms"`
	Error      string      `json:"error,omitempty"`
}

// JobList lists all jobs
func JobList(namespace string) ([]JobInfo, error) {
	args := []string{"job", "status", "-json"}
	if namespace != "" {
		args = append(args, "-namespace", namespace)
	}

	cmd := exec.Command("nomad", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("nomad job list failed: %v, output: %s", err, string(output))
	}

	var jobs []JobInfo
	if err := json.Unmarshal(output, &jobs); err != nil {
		return nil, fmt.Errorf("failed to parse job list: %v", err)
	}

	return jobs, nil
}

// JobRun runs a job
func JobRun(jobFile, namespace string) (NomadResult, error) {
	start := time.Now()

	if jobFile == "" {
		return NomadResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("job file required")
	}

	args := []string{"job", "run"}
	if namespace != "" {
		args = append(args, "-namespace", namespace)
	}
	args = append(args, jobFile)

	cmd := exec.Command("nomad", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := NomadResult{
		Changed:    true,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("nomad job run failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Job submitted successfully"
	return result, nil
}

// JobStop stops a job
func JobStop(jobID, namespace string) (NomadResult, error) {
	start := time.Now()

	if jobID == "" {
		return NomadResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("job ID required")
	}

	args := []string{"job", "stop"}
	if namespace != "" {
		args = append(args, "-namespace", namespace)
	}
	args = append(args, jobID)

	cmd := exec.Command("nomad", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := NomadResult{
		Changed:    true,
		ID:         jobID,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("nomad job stop failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Job stopped successfully"
	return result, nil
}

// AllocList lists allocations
func AllocList(jobID, namespace string) ([]AllocInfo, error) {
	args := []string{"alloc", "status", "-json"}
	if namespace != "" {
		args = append(args, "-namespace", namespace)
	}
	if jobID != "" {
		args = append(args, "-job", jobID)
	}

	cmd := exec.Command("nomad", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("nomad alloc list failed: %v, output: %s", err, string(output))
	}

	var allocs []AllocInfo
	if err := json.Unmarshal(output, &allocs); err != nil {
		return nil, fmt.Errorf("failed to parse alloc list: %v", err)
	}

	return allocs, nil
}

// NodeList lists all nodes
func NodeList() ([]NodeInfo, error) {
	cmd := exec.Command("nomad", "node", "status", "-json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("nomad node list failed: %v, output: %s", err, string(output))
	}

	var nodes []NodeInfo
	if err := json.Unmarshal(output, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse node list: %v", err)
	}

	return nodes, nil
}

// NodeDrain toggles drain mode on a node
func NodeDrain(nodeID string, enable bool) (NomadResult, error) {
	start := time.Now()

	if nodeID == "" {
		return NomadResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("node ID required")
	}

	args := []string{"node", "drain"}
	if enable {
		args = append(args, "-enable", "-yes")
	} else {
		args = append(args, "-disable")
	}
	args = append(args, nodeID)

	cmd := exec.Command("nomad", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := NomadResult{
		Changed:    true,
		ID:         nodeID,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("nomad node drain failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Node drain updated successfully"
	return result, nil
}
