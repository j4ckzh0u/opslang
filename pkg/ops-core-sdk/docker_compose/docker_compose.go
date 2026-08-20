// Package docker_compose provides Docker Compose management operations.
package docker_compose

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ActionResult represents the result of a compose action.
type ActionResult struct {
	Project    string `json:"project,omitempty"`
	Changed    bool   `json:"changed"`
	Action     string `json:"action"`
	DurationMs int64  `json:"duration_ms"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ServiceStatus represents the status of a compose service.
type ServiceStatus struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Running bool   `json:"running"`
	Ports   string `json:"ports,omitempty"`
}

// StatusResult represents the result of getting compose status.
type StatusResult struct {
	Project  string          `json:"project,omitempty"`
	Services []ServiceStatus `json:"services"`
}

func findCompose() (string, error) {
	// Try `docker compose` (v2) first
	if _, err := exec.LookPath("docker"); err == nil {
		cmd := exec.Command("docker", "compose", "version")
		if err := cmd.Run(); err == nil {
			return "docker", nil
		}
	}
	// Fallback to `docker-compose` (v1)
	if path, err := exec.LookPath("docker-compose"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("neither 'docker compose' nor 'docker-compose' found in PATH")
}

func runCompose(args ...string) (string, error) {
	bin, err := findCompose()
	if err != nil {
		return "", err
	}
	var cmdArgs []string
	if bin == "docker" {
		cmdArgs = append([]string{"compose"}, args...)
	} else {
		cmdArgs = args
	}
	cmd := exec.Command(bin, cmdArgs...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Up starts compose services (idempotent - docker compose up -d).
func Up(projectDir string) (*ActionResult, error) {
	start := time.Now()
	args := []string{"up", "-d"}
	if projectDir != "" {
		args = append([]string{"-f", projectDir + "/docker-compose.yml"}, args...)
	}
	output, err := runCompose(args...)
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "up",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     output,
			Error:      err.Error(),
		}, fmt.Errorf("docker compose up: %s", output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "up",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(output),
	}, nil
}

// Down stops and removes compose services.
func Down(projectDir string) (*ActionResult, error) {
	start := time.Now()
	args := []string{"down"}
	if projectDir != "" {
		args = append([]string{"-f", projectDir + "/docker-compose.yml"}, args...)
	}
	output, err := runCompose(args...)
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "down",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     output,
			Error:      err.Error(),
		}, fmt.Errorf("docker compose down: %s", output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "down",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(output),
	}, nil
}

// Restart restarts compose services.
func Restart(projectDir string) (*ActionResult, error) {
	start := time.Now()
	args := []string{"restart"}
	if projectDir != "" {
		args = append([]string{"-f", projectDir + "/docker-compose.yml"}, args...)
	}
	output, err := runCompose(args...)
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "restart",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     output,
			Error:      err.Error(),
		}, fmt.Errorf("docker compose restart: %s", output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "restart",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(output),
	}, nil
}

// Pull pulls compose service images.
func Pull(projectDir string) (*ActionResult, error) {
	start := time.Now()
	args := []string{"pull"}
	if projectDir != "" {
		args = append([]string{"-f", projectDir + "/docker-compose.yml"}, args...)
	}
	output, err := runCompose(args...)
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "pull",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     output,
			Error:      err.Error(),
		}, fmt.Errorf("docker compose pull: %s", output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "pull",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(output),
	}, nil
}

// Status returns the status of compose services.
func Status(projectDir string) (*StatusResult, error) {
	args := []string{"ps", "--format", "{{.Name}}\t{{.State}}\t{{.Ports}}"}
	if projectDir != "" {
		args = append([]string{"-f", projectDir + "/docker-compose.yml"}, args...)
	}
	output, err := runCompose(args...)
	if err != nil {
		return nil, fmt.Errorf("docker compose ps: %s", output)
	}

	result := &StatusResult{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		svc := ServiceStatus{Name: parts[0]}
		if len(parts) >= 2 {
			svc.State = parts[1]
			svc.Running = strings.Contains(strings.ToLower(parts[1]), "running") || strings.Contains(strings.ToLower(parts[1]), "up")
		}
		if len(parts) >= 3 {
			svc.Ports = parts[2]
		}
		result.Services = append(result.Services, svc)
	}

	return result, nil
}

// Build builds compose service images.
func Build(projectDir string) (*ActionResult, error) {
	start := time.Now()
	args := []string{"build"}
	if projectDir != "" {
		args = append([]string{"-f", projectDir + "/docker-compose.yml"}, args...)
	}
	output, err := runCompose(args...)
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "build",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     output,
			Error:      err.Error(),
		}, fmt.Errorf("docker compose build: %s", output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "build",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(output),
	}, nil
}

// Logs retrieves logs from compose services.
func Logs(projectDir string, tail int) (string, error) {
	args := []string{"logs"}
	if tail > 0 {
		args = append(args, fmt.Sprintf("--tail=%d", tail))
	}
	if projectDir != "" {
		args = append([]string{"-f", projectDir + "/docker-compose.yml"}, args...)
	}
	output, err := runCompose(args...)
	if err != nil {
		return "", fmt.Errorf("docker compose logs: %s", output)
	}
	return output, nil
}
