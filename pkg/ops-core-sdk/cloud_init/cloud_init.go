// Package cloud_init provides cloud-init module management operations.
package cloud_init

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// StatusResult represents the result of getting cloud-init status.
type StatusResult struct {
	Status    string `json:"status"`
	Available bool   `json:"available"`
	Datasource string `json:"datasource,omitempty"`
	Stages    []string `json:"stages,omitempty"`
}

// ActionResult represents the result of a cloud-init action.
type ActionResult struct {
	Changed    bool   `json:"changed"`
	Action     string `json:"action"`
	DurationMs int64  `json:"duration_ms"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ModulesResult represents the result of listing cloud-init modules.
type ModulesResult struct {
	Init    []string `json:"init,omitempty"`
	Config  []string `json:"config,omitempty"`
	Final   []string `json:"final,omitempty"`
}

// Status returns the cloud-init status.
func Status() (*StatusResult, error) {
	cmd := exec.Command("cloud-init", "status")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &StatusResult{Status: "unknown", Available: false}, nil
	}

	result := &StatusResult{Available: true}
	output := string(out)

	if strings.Contains(output, "status: done") {
		result.Status = "done"
	} else if strings.Contains(output, "status: running") {
		result.Status = "running"
	} else if strings.Contains(output, "status: not run") {
		result.Status = "not_run"
	} else if strings.Contains(output, "status: error") {
		result.Status = "error"
	} else {
		result.Status = "unknown"
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "DataSource") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result.Datasource = strings.TrimSpace(parts[1])
			}
		}
	}

	return result, nil
}

// Modules returns the cloud-init modules configured to run.
func Modules() (*ModulesResult, error) {
	cmd := exec.Command("cloud-init", "modules", "--list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cloud-init modules: %s", string(out))
	}

	result := &ModulesResult{}
	currentStage := ""
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "init") {
			currentStage = "init"
			continue
		}
		if strings.Contains(line, "config") {
			currentStage = "config"
			continue
		}
		if strings.Contains(line, "final") {
			currentStage = "final"
			continue
		}
		if line == "" || !strings.HasPrefix(line, "-") {
			continue
		}
		mod := strings.TrimPrefix(line, "- ")
		mod = strings.TrimSpace(mod)
		switch currentStage {
		case "init":
			result.Init = append(result.Init, mod)
		case "config":
			result.Config = append(result.Config, mod)
		case "final":
			result.Final = append(result.Final, mod)
		}
	}

	return result, nil
}

// Clean cleans cloud-init artifacts and allows re-running.
func Clean(removeLogs bool) (*ActionResult, error) {
	start := time.Now()
	args := []string{"clean"}
	if removeLogs {
		args = append(args, "--logs")
	}
	cmd := exec.Command("cloud-init", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "clean",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(out),
			Error:      err.Error(),
		}, fmt.Errorf("cloud-init clean: %s", string(out))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "clean",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(string(out)),
	}, nil
}

// Init runs cloud-init initialization.
func Init() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("cloud-init", "init")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "init",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(out),
			Error:      err.Error(),
		}, fmt.Errorf("cloud-init init: %s", string(out))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "init",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(string(out)),
	}, nil
}
