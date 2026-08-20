// Package docker_container provides Docker container management via docker CLI.
// Equivalent to Ansible's docker_container module.
package docker_container

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Result represents a container operation result.
type Result struct {
	Status    string `json:"status"`
	Changed   bool   `json:"changed"`
	Container string `json:"container"`
	ID        string `json:"id,omitempty"`
	State     string `json:"state,omitempty"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ContainerInfo represents Docker container metadata.
type ContainerInfo struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	State   string   `json:"state"`
	Status  string   `json:"status"`
	Ports   []string `json:"ports,omitempty"`
	Created string   `json:"created"`
}

type dockerContainerJSON struct {
	ID      string `json:"Id"`
	Name    string `json:"Name"`
	Image   string `json:"Image"`
	State   string `json:"State"`
	Created string `json:"Created"`
}

func findDocker() string {
	if p, err := exec.LookPath("docker"); err == nil {
		return p
	}
	return ""
}

// Start starts a container.
func Start(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}

	out, err := exec.Command(docker, "start", name).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Container: name, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Container: name, State: "running", Output: outStr}, nil
}

// Stop stops a container.
func Stop(name string, timeout int) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}

	args := []string{"stop"}
	if timeout > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", timeout))
	}
	args = append(args, name)

	out, err := exec.Command(docker, args...).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Container: name, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Container: name, State: "stopped", Output: outStr}, nil
}

// Remove removes a container.
func Remove(name string, force bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}

	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	out, err := exec.Command(docker, args...).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Container: name, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Container: name, Output: outStr}, nil
}

// Restart restarts a container.
func Restart(name string, timeout int) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}

	args := []string{"restart"}
	if timeout > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", timeout))
	}
	args = append(args, name)

	out, err := exec.Command(docker, args...).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Container: name, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Container: name, State: "running", Output: outStr}, nil
}

// Pause pauses a running container.
func Pause(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}

	out, err := exec.Command(docker, "pause", name).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Container: name, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Container: name, State: "paused", Output: outStr}, nil
}

// Unpause resumes a paused container.
func Unpause(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}

	out, err := exec.Command(docker, "unpause", name).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Container: name, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Container: name, State: "running", Output: outStr}, nil
}

// Inspect returns container metadata.
func Inspect(name string) (ContainerInfo, error) {
	if name == "" {
		return ContainerInfo{State: "failed"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return ContainerInfo{State: "failed"}, fmt.Errorf("docker not found")
	}

	out, err := exec.Command(docker, "inspect", name).CombinedOutput()
	if err != nil {
		return ContainerInfo{Name: name, State: "not_found"}, nil
	}

	var containers []dockerContainerJSON
	if err := json.Unmarshal(out, &containers); err != nil || len(containers) == 0 {
		return ContainerInfo{Name: name, State: "not_found"}, nil
	}

	c := containers[0]
	cName := strings.TrimPrefix(c.Name, "/")
	shortID := c.ID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}
	return ContainerInfo{
		ID:      shortID,
		Name:    cName,
		Image:   c.Image,
		State:   c.State,
		Created: c.Created,
		Status:  c.State,
	}, nil
}

// List returns all containers.
func List(all bool) ([]ContainerInfo, error) {
	docker := findDocker()
	if docker == "" {
		return nil, fmt.Errorf("docker not found")
	}

	args := []string{"ps", "--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.CreatedAt}}"}
	if all {
		args = append(args, "-a")
	}

	out, err := exec.Command(docker, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker ps: %w", err)
	}

	var containers []ContainerInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 6)
		if len(parts) < 4 {
			continue
		}
		c := ContainerInfo{
			ID:     parts[0],
			Name:   parts[1],
			Image:  parts[2],
			State:  parts[3],
			Status: parts[3],
		}
		if len(parts) >= 6 {
			c.Created = parts[5]
		}
		containers = append(containers, c)
	}
	return containers, nil
}

// Logs returns container logs.
func Logs(name string, tail string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return "", fmt.Errorf("docker not found")
	}

	args := []string{"logs"}
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	args = append(args, name)

	out, err := exec.Command(docker, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker logs: %w", err)
	}
	return string(out), nil
}
