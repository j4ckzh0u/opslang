// Package podman provides Podman container management operations.
// Equivalent to Ansible containers.podman collection.
package podman

import (
	"fmt"
	"os/exec"
	"strings"
)

// ContainerResult represents a container operation result.
type ContainerResult struct {
	Status      string `json:"status"`
	Changed     bool   `json:"changed"`
	ContainerID string `json:"container_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ContainerInfo represents container details.
type ContainerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Ports   string `json:"ports,omitempty"`
	Created string `json:"created,omitempty"`
}

// ImageResult represents an image operation result.
type ImageResult struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	ImageID string `json:"image_id,omitempty"`
	Tag     string `json:"tag,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PodResult represents a pod operation result.
type PodResult struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	PodID   string `json:"pod_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

func findPodman() (string, error) {
	if p, err := exec.LookPath("podman"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("podman not found")
}

func runCmd(args ...string) (string, error) {
	podman, err := findPodman()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(podman, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Run creates and starts a container.
func Run(image string, name string, command string) (ContainerResult, error) {
	if image == "" {
		return ContainerResult{Status: "failed", Error: "image is required"}, fmt.Errorf("image is required")
	}
	args := []string{"run", "-d"}
	if name != "" {
		args = append(args, "--name", name)
	}
	args = append(args, image)
	if command != "" {
		args = append(args, command)
	}
	out, err := runCmd(args...)
	if err != nil {
		return ContainerResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman run: %v", err)}, err
	}
	return ContainerResult{Status: "success", Changed: true, ContainerID: out, Name: name, Output: out}, nil
}

// Stop stops a container.
func Stop(nameOrID string, timeout int) (ContainerResult, error) {
	if nameOrID == "" {
		return ContainerResult{Status: "failed", Error: "name or ID is required"}, fmt.Errorf("name or ID is required")
	}
	args := []string{"stop"}
	if timeout > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", timeout))
	}
	args = append(args, nameOrID)
	out, err := runCmd(args...)
	if err != nil {
		return ContainerResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman stop: %v", err)}, err
	}
	return ContainerResult{Status: "success", Changed: true, ContainerID: out, Name: nameOrID, Output: out}, nil
}

// Start starts a stopped container.
func Start(nameOrID string) (ContainerResult, error) {
	if nameOrID == "" {
		return ContainerResult{Status: "failed", Error: "name or ID is required"}, fmt.Errorf("name or ID is required")
	}
	out, err := runCmd("start", nameOrID)
	if err != nil {
		return ContainerResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman start: %v", err)}, err
	}
	return ContainerResult{Status: "success", Changed: true, ContainerID: out, Name: nameOrID, Output: out}, nil
}

// Remove removes a container.
func Remove(nameOrID string, force bool) (ContainerResult, error) {
	if nameOrID == "" {
		return ContainerResult{Status: "failed", Error: "name or ID is required"}, fmt.Errorf("name or ID is required")
	}
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, nameOrID)
	out, err := runCmd(args...)
	if err != nil {
		return ContainerResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman rm: %v", err)}, err
	}
	return ContainerResult{Status: "success", Changed: true, ContainerID: out, Name: nameOrID, Output: out}, nil
}

// ListContainers lists containers.
func ListContainers(all bool) ([]ContainerInfo, error) {
	args := []string{"ps", "--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.State}}|{{.Status}}|{{.Ports}}|{{.Created}}"}
	if all {
		args = append(args, "-a")
	}
	out, err := runCmd(args...)
	if err != nil {
		return nil, fmt.Errorf("podman ps: %w", err)
	}
	var results []ContainerInfo
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			info := ContainerInfo{
				ID:     parts[0],
				Name:   parts[1],
				Image:  parts[2],
				State:  parts[3],
				Status: parts[4],
			}
			if len(parts) >= 6 {
				info.Ports = parts[5]
			}
			if len(parts) >= 7 {
				info.Created = parts[6]
			}
			results = append(results, info)
		}
	}
	return results, nil
}

// Inspect returns detailed container information.
func Inspect(nameOrID string) (map[string]interface{}, error) {
	if nameOrID == "" {
		return nil, fmt.Errorf("name or ID is required")
	}
	out, err := runCmd("inspect", nameOrID)
	if err != nil {
		return nil, fmt.Errorf("podman inspect: %w", err)
	}
	// Parse JSON output
	var result map[string]interface{}
	// For simplicity, return raw output as a map
	result = map[string]interface{}{
		"raw": out,
	}
	return result, nil
}

// Pull pulls an image.
func Pull(image string) (ImageResult, error) {
	if image == "" {
		return ImageResult{Status: "failed", Error: "image is required"}, fmt.Errorf("image is required")
	}
	out, err := runCmd("pull", image)
	if err != nil {
		return ImageResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman pull: %v", err)}, err
	}
	return ImageResult{Status: "success", Changed: true, Tag: image, Output: out}, nil
}

// ListImages lists images.
func ListImages() ([]map[string]string, error) {
	out, err := runCmd("images", "--format", "{{.ID}}|{{.Repository}}|{{.Tag}}|{{.Size}}|{{.Created}}")
	if err != nil {
		return nil, fmt.Errorf("podman images: %w", err)
	}
	var results []map[string]string
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			img := map[string]string{
				"id":         parts[0],
				"repository": parts[1],
				"tag":        parts[2],
				"size":       parts[3],
			}
			if len(parts) >= 5 {
				img["created"] = parts[4]
			}
			results = append(results, img)
		}
	}
	return results, nil
}

// RemoveImage removes an image.
func RemoveImage(imageID string, force bool) (ImageResult, error) {
	if imageID == "" {
		return ImageResult{Status: "failed", Error: "image ID is required"}, fmt.Errorf("image ID is required")
	}
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, imageID)
	out, err := runCmd(args...)
	if err != nil {
		return ImageResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman rmi: %v", err)}, err
	}
	return ImageResult{Status: "success", Changed: true, ImageID: out, Output: out}, nil
}

// CreatePod creates a pod.
func CreatePod(name string) (PodResult, error) {
	if name == "" {
		return PodResult{Status: "failed", Error: "pod name is required"}, fmt.Errorf("pod name is required")
	}
	out, err := runCmd("pod", "create", "--name", name)
	if err != nil {
		return PodResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman pod create: %v", err)}, err
	}
	return PodResult{Status: "success", Changed: true, PodID: out, Name: name, Output: out}, nil
}

// StopPod stops a pod.
func StopPod(nameOrID string) (PodResult, error) {
	if nameOrID == "" {
		return PodResult{Status: "failed", Error: "pod name or ID is required"}, fmt.Errorf("pod name or ID is required")
	}
	out, err := runCmd("pod", "stop", nameOrID)
	if err != nil {
		return PodResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman pod stop: %v", err)}, err
	}
	return PodResult{Status: "success", Changed: true, PodID: out, Name: nameOrID, Output: out}, nil
}

// RemovePod removes a pod.
func RemovePod(nameOrID string, force bool) (PodResult, error) {
	if nameOrID == "" {
		return PodResult{Status: "failed", Error: "pod name or ID is required"}, fmt.Errorf("pod name or ID is required")
	}
	args := []string{"pod", "rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, nameOrID)
	out, err := runCmd(args...)
	if err != nil {
		return PodResult{Status: "failed", Output: out, Error: fmt.Sprintf("podman pod rm: %v", err)}, err
	}
	return PodResult{Status: "success", Changed: true, PodID: out, Name: nameOrID, Output: out}, nil
}

// ListPods lists pods.
func ListPods() ([]map[string]string, error) {
	out, err := runCmd("pod", "ps", "--format", "{{.Id}}|{{.Name}}|{{.Status}}|{{.Created}}")
	if err != nil {
		return nil, fmt.Errorf("podman pod ps: %w", err)
	}
	var results []map[string]string
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			pod := map[string]string{
				"id":     parts[0],
				"name":   parts[1],
				"status": parts[2],
			}
			if len(parts) >= 4 {
				pod["created"] = parts[3]
			}
			results = append(results, pod)
		}
	}
	return results, nil
}
