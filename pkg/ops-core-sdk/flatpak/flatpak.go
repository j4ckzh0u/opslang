// Package flatpak provides Flatpak package management operations.
package flatpak

import (
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a flatpak operation.
type ActionResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// FlatpakInfo represents information about a flatpak package.
type FlatpakInfo struct {
	Name       string `json:"name"`
	AppID      string `json:"app_id"`
	Version    string `json:"version"`
	Branch     string `json:"branch"`
	Origin     string `json:"origin"`
	Installation string `json:"installation"`
}

// ListResult represents the result of listing flatpaks.
type ListResult struct {
	Apps []FlatpakInfo `json:"apps"`
}

// Install installs a flatpak package.
func Install(name string, from string, user bool) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("flatpak name is required")
	}

	args := []string{"install", "-y"}
	if user {
		args = append(args, "--user")
	} else {
		args = append(args, "--system")
	}
	if from != "" {
		args = append(args, from)
	}
	args = append(args, name)

	out, err := exec.Command("flatpak", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("flatpak install failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Remove removes a flatpak package.
func Remove(name string, user bool) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("flatpak name is required")
	}

	args := []string{"uninstall", "-y"}
	if user {
		args = append(args, "--user")
	} else {
		args = append(args, "--system")
	}
	args = append(args, name)

	out, err := exec.Command("flatpak", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("flatpak remove failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Update updates a flatpak package.
func Update(name string, user bool) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("flatpak name is required")
	}

	args := []string{"update", "-y"}
	if user {
		args = append(args, "--user")
	} else {
		args = append(args, "--system")
	}
	args = append(args, name)

	out, err := exec.Command("flatpak", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("flatpak update failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// List lists installed flatpaks.
func List(user bool) (ListResult, error) {
	args := []string{"list", "--app", "--columns=application,version,branch,origin,installation"}
	if user {
		args = append(args, "--user")
	} else {
		args = append(args, "--system")
	}

	out, err := exec.Command("flatpak", args...).CombinedOutput()
	if err != nil {
		return ListResult{}, fmt.Errorf("flatpak list failed: %w (output: %s)", err, string(out))
	}

	result := ListResult{Apps: make([]FlatpakInfo, 0)}
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		// Parse: Application Version Branch Origin Installation
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			app := FlatpakInfo{
				AppID:   fields[0],
				Version: fields[1],
				Branch:  fields[2],
				Origin:  fields[3],
			}
			if len(fields) >= 5 {
				app.Installation = fields[4]
			}
			// Try to get friendly name
			app.Name = fields[0]
			result.Apps = append(result.Apps, app)
		}
	}
	return result, nil
}

// Info gets information about a specific flatpak.
func Info(name string, user bool) (FlatpakInfo, error) {
	if name == "" {
		return FlatpakInfo{}, fmt.Errorf("flatpak name is required")
	}

	args := []string{"info", "--show-commit"}
	if user {
		args = append(args, "--user")
	} else {
		args = append(args, "--system")
	}
	args = append(args, name)

	out, err := exec.Command("flatpak", args...).CombinedOutput()
	if err != nil {
		return FlatpakInfo{}, fmt.Errorf("flatpak info failed: %w (output: %s)", err, string(out))
	}

	info := FlatpakInfo{Name: name, AppID: name}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ID:") {
			info.AppID = strings.TrimSpace(strings.TrimPrefix(line, "ID:"))
		} else if strings.HasPrefix(line, "Version:") {
			info.Version = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		} else if strings.HasPrefix(line, "Branch:") {
			info.Branch = strings.TrimSpace(strings.TrimPrefix(line, "Branch:"))
		} else if strings.HasPrefix(line, "Origin:") {
			info.Origin = strings.TrimSpace(strings.TrimPrefix(line, "Origin:"))
		}
	}
	return info, nil
}

// Run runs a flatpak application.
func Run(name string, args []string, user bool) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("flatpak name is required")
	}

	cmdArgs := []string{"run"}
	if user {
		cmdArgs = append(cmdArgs, "--user")
	} else {
		cmdArgs = append(cmdArgs, "--system")
	}
	cmdArgs = append(cmdArgs, name)
	cmdArgs = append(cmdArgs, args...)

	out, err := exec.Command("flatpak", cmdArgs...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("flatpak run failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: false, Success: true}, nil
}

// Repair repairs a flatpak installation.
func Repair(user bool) (ActionResult, error) {
	args := []string{"repair"}
	if user {
		args = append(args, "--user")
	} else {
		args = append(args, "--system")
	}

	out, err := exec.Command("flatpak", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Success: false, Error: string(out)}, fmt.Errorf("flatpak repair failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Changed: true, Success: true}, nil
}
