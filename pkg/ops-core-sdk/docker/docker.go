// Package docker provides Docker container and image management operations.
package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ContainerInfo represents a Docker container.
type ContainerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	State   string `json:"state"`
	Created string `json:"created"`
}

// ImageInfo represents a Docker image.
type ImageInfo struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Size       string `json:"size"`
	Created    string `json:"created"`
}

// ContainerListResult is returned by ContainerList.
type ContainerListResult struct {
	Containers []ContainerInfo `json:"containers"`
}

// ImageListResult is returned by ImageList.
type ImageListResult struct {
	Images []ImageInfo `json:"images"`
}

// ContainerExistsResult is returned by ContainerExists.
type ContainerExistsResult struct {
	Exists bool   `json:"exists"`
	ID     string `json:"id,omitempty"`
}

// RunResult is returned by ContainerRun.
type RunResult struct {
	Changed   bool   `json:"changed"`
	ID        string `json:"id"`
	Error     string `json:"error,omitempty"`
}

// RemoveResult is returned by ContainerRemove/ImageRemove.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// PullResult is returned by ImagePull.
type PullResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ContainerList lists all Docker containers.
func ContainerList(all bool) (ContainerListResult, error) {
	args := []string{"ps", "--format", "{{json .}}"}
	if all {
		args = append(args, "-a")
	}

	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return ContainerListResult{}, fmt.Errorf("docker ps failed: %w: %s", err, stderr.String())
	}

	var containers []ContainerInfo
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var c ContainerInfo
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			continue
		}
		containers = append(containers, c)
	}

	return ContainerListResult{Containers: containers}, nil
}

// ContainerExists checks if a container exists.
func ContainerExists(name string) (ContainerExistsResult, error) {
	cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.ID}}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return ContainerExistsResult{}, fmt.Errorf("docker ps failed: %w", err)
	}

	id := strings.TrimSpace(stdout.String())
	if id == "" {
		return ContainerExistsResult{Exists: false}, nil
	}

	return ContainerExistsResult{Exists: true, ID: id}, nil
}

// ContainerRun runs a Docker container.
func ContainerRun(name, image string, opts map[string]string) (RunResult, error) {
	if image == "" {
		return RunResult{Error: "image is required"}, fmt.Errorf("image is required")
	}

	// Check if container already exists
	exists, err := ContainerExists(name)
	if err == nil && exists.Exists {
		return RunResult{Changed: false, ID: exists.ID}, nil
	}

	args := []string{"run", "-d"}
	if name != "" {
		args = append(args, "--name", name)
	}

	// Handle common options
	if opts != nil {
		if ports, ok := opts["ports"]; ok && ports != "" {
			for _, p := range strings.Split(ports, ",") {
				args = append(args, "-p", strings.TrimSpace(p))
			}
		}
		if env, ok := opts["env"]; ok && env != "" {
			for _, e := range strings.Split(env, ",") {
				args = append(args, "-e", strings.TrimSpace(e))
			}
		}
		if volumes, ok := opts["volumes"]; ok && volumes != "" {
			for _, v := range strings.Split(volumes, ",") {
				args = append(args, "-v", strings.TrimSpace(v))
			}
		}
	}

	args = append(args, image)

	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return RunResult{Error: stderr.String()}, fmt.Errorf("docker run failed: %w: %s", err, stderr.String())
	}

	id := strings.TrimSpace(stdout.String())
	return RunResult{Changed: true, ID: id}, nil
}

// ContainerStop stops a Docker container.
func ContainerStop(name string) (RemoveResult, error) {
	cmd := exec.Command("docker", "stop", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return RemoveResult{Error: stderr.String()}, fmt.Errorf("docker stop failed: %w: %s", err, stderr.String())
	}

	return RemoveResult{Changed: true}, nil
}

// ContainerRemove removes a Docker container.
func ContainerRemove(name string, force bool) (RemoveResult, error) {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	cmd := exec.Command("docker", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return RemoveResult{Error: stderr.String()}, fmt.Errorf("docker rm failed: %w: %s", err, stderr.String())
	}

	return RemoveResult{Changed: true}, nil
}

// ImageList lists Docker images.
func ImageList() (ImageListResult, error) {
	cmd := exec.Command("docker", "images", "--format", "{{json .}}")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return ImageListResult{}, fmt.Errorf("docker images failed: %w: %s", err, stderr.String())
	}

	var images []ImageInfo
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var img ImageInfo
		if err := json.Unmarshal([]byte(line), &img); err != nil {
			continue
		}
		images = append(images, img)
	}

	return ImageListResult{Images: images}, nil
}

// ImagePull pulls a Docker image.
func ImagePull(image string) (PullResult, error) {
	if image == "" {
		return PullResult{Error: "image is required"}, fmt.Errorf("image is required")
	}

	cmd := exec.Command("docker", "pull", image)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return PullResult{Error: stderr.String()}, fmt.Errorf("docker pull failed: %w: %s", err, stderr.String())
	}

	return PullResult{Changed: true}, nil
}

// ImageRemove removes a Docker image.
func ImageRemove(image string, force bool) (RemoveResult, error) {
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, image)

	cmd := exec.Command("docker", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return RemoveResult{Error: stderr.String()}, fmt.Errorf("docker rmi failed: %w: %s", err, stderr.String())
	}

	return RemoveResult{Changed: true}, nil
}
