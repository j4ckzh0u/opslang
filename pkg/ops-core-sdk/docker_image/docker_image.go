// Package docker_image provides Docker image management via docker CLI.
// Equivalent to Ansible's docker_image module.
package docker_image

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Result represents a docker image operation result.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Image   string `json:"image"`
	Tag     string `json:"tag,omitempty"`
	ID      string `json:"id,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ImageInfo represents Docker image metadata.
type ImageInfo struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Size    int64    `json:"size"`
	Created string   `json:"created"`
	Status  string   `json:"status"`
}

type dockerImageJSON struct {
	ID       string   `json:"Id"`
	RepoTags []string `json:"RepoTags"`
	Size     int64    `json:"Size"`
	Created  string   `json:"Created"`
}

func findDocker() string {
	if p, err := exec.LookPath("docker"); err == nil {
		return p
	}
	return ""
}

// Pull pulls a Docker image.
func Pull(name string, tag string, force bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}
	if tag == "" {
		tag = "latest"
	}

	ref := name + ":" + tag
	out, err := exec.Command(docker, "pull", ref).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Image: ref, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Image: name, Tag: tag, Output: outStr}, nil
}

// Build builds a Docker image from a Dockerfile.
func Build(path string, name string, tag string, dockerfile string) (Result, error) {
	if path == "" || name == "" {
		return Result{Status: "failed", Error: "path and name are required"}, fmt.Errorf("path and name are required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}
	if tag == "" {
		tag = "latest"
	}

	args := []string{"build", "-t", name + ":" + tag}
	if dockerfile != "" {
		args = append(args, "-f", dockerfile)
	}
	args = append(args, path)

	out, err := exec.Command(docker, args...).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Image: name, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Image: name, Tag: tag, Output: outStr}, nil
}

// Remove removes a Docker image.
func Remove(name string, tag string, force bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}
	if tag == "" {
		tag = "latest"
	}

	ref := name + ":" + tag
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, ref)

	out, err := exec.Command(docker, args...).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Image: ref, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Image: name, Tag: tag, Output: outStr}, nil
}

// Tag tags a Docker image.
func Tag(source string, target string) (Result, error) {
	if source == "" || target == "" {
		return Result{Status: "failed", Error: "source and target are required"}, fmt.Errorf("source and target are required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}

	out, err := exec.Command(docker, "tag", source, target).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Image: source, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Image: source, Output: outStr}, nil
}

// Inspect returns metadata for a Docker image.
func Inspect(name string) (ImageInfo, error) {
	if name == "" {
		return ImageInfo{Status: "failed"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return ImageInfo{Status: "failed"}, fmt.Errorf("docker not found")
	}

	out, err := exec.Command(docker, "image", "inspect", name).CombinedOutput()
	if err != nil {
		return ImageInfo{ID: name, Status: "not_found"}, nil
	}

	var images []dockerImageJSON
	if err := json.Unmarshal(out, &images); err != nil || len(images) == 0 {
		return ImageInfo{ID: name, Status: "not_found"}, nil
	}

	img := images[0]
	shortID := img.ID
	if len(shortID) > 19 {
		shortID = shortID[:19]
	}
	return ImageInfo{
		ID:      shortID,
		Tags:    img.RepoTags,
		Size:    img.Size,
		Created: img.Created,
		Status:  "present",
	}, nil
}

// List returns all Docker images.
func List() ([]ImageInfo, error) {
	docker := findDocker()
	if docker == "" {
		return nil, fmt.Errorf("docker not found")
	}

	out, err := exec.Command(docker, "image", "ls", "--format", "{{.ID}}\t{{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker image ls: %w", err)
	}

	var images []ImageInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) < 2 {
			continue
		}
		img := ImageInfo{ID: parts[0], Status: "present"}
		if len(parts) >= 2 {
			img.Tags = []string{parts[1]}
		}
		if len(parts) >= 4 {
			img.Created = parts[3]
		}
		images = append(images, img)
	}
	return images, nil
}

// Push pushes a Docker image to a registry.
func Push(name string, tag string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	docker := findDocker()
	if docker == "" {
		return Result{Status: "failed", Error: "docker not found"}, fmt.Errorf("docker not found")
	}
	if tag == "" {
		tag = "latest"
	}

	ref := name + ":" + tag
	out, err := exec.Command(docker, "push", ref).CombinedOutput()
	outStr := string(out)
	if err != nil {
		return Result{Status: "failed", Image: ref, Output: outStr, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Image: name, Tag: tag, Output: outStr}, nil
}
