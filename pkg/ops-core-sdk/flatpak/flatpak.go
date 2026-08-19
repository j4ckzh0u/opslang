// Package flatpak manages Flatpak applications.
// Equivalent to community.general.flatpak module.
package flatpak

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Package string `json:"package,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Status string   `json:"status"`
	Apps   []string `json:"apps"`
	Error  string   `json:"error,omitempty"`
}

// InfoResult is returned by Info.
type InfoResult struct {
	Status  string `json:"status"`
	Package string `json:"package"`
	Version string `json:"version,omitempty"`
	Origin  string `json:"origin,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Install installs a Flatpak application.
func Install(name string, from string, user bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	if from == "" {
		from = "flathub"
	}

	args := []string{"install", "-y", from, name}
	if user {
		args = append([]string{"install", "-y", "--user", from, name}, args[4:]...)
		args = []string{"install", "-y", "--user", from, name}
	}

	cmd := exec.Command("flatpak", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("flatpak install: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}, nil
}

// Remove removes a Flatpak application.
func Remove(name string, user bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}

	args := []string{"uninstall", "-y", name}
	if user {
		args = []string{"uninstall", "-y", "--user", name}
	}

	cmd := exec.Command("flatpak", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("flatpak uninstall: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}, nil
}

// List lists installed Flatpak applications.
func List(user bool) (ListResult, error) {
	args := []string{"list", "--app", "--columns=application"}
	if user {
		args = []string{"list", "--user", "--app", "--columns=application"}
	}

	cmd := exec.Command("flatpak", args...)
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("flatpak list: %v", err)}, err
	}

	apps := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			apps = append(apps, line)
		}
	}
	return ListResult{Status: "success", Apps: apps}, nil
}

// Update updates a Flatpak application.
func Update(name string, user bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}

	args := []string{"update", "-y", name}
	if user {
		args = []string{"update", "-y", "--user", name}
	}

	cmd := exec.Command("flatpak", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("flatpak update: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}, nil
}

// Info returns information about a Flatpak application.
func Info(name string, user bool) (InfoResult, error) {
	if name == "" {
		return InfoResult{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}

	args := []string{"info", name}
	if user {
		args = []string{"info", "--user", name}
	}

	cmd := exec.Command("flatpak", args...)
	out, err := cmd.Output()
	if err != nil {
		return InfoResult{Status: "failed", Error: fmt.Sprintf("flatpak info: %v", err)}, err
	}

	output := strings.TrimSpace(string(out))
	version := ""
	origin := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "version:") {
			version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		}
		if strings.HasPrefix(line, "origin:") {
			origin = strings.TrimSpace(strings.TrimPrefix(line, "origin:"))
		}
	}
	return InfoResult{Status: "success", Package: name, Version: version, Origin: origin, Output: output}, nil
}

// Run runs a Flatpak application.
func Run(name string, runArgs []string, user bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}

	args := []string{"run"}
	if user {
		args = append(args, "--user")
	}
	args = append(args, name)
	args = append(args, runArgs...)

	cmd := exec.Command("flatpak", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("flatpak run: %v", err)}, err
	}
	return Result{Status: "success", Package: name, Output: output}, nil
}

// Repair repairs a Flatpak installation.
func Repair(user bool) (Result, error) {
	args := []string{"repair", "--force"}
	if user {
		args = append(args, "--user")
	}

	cmd := exec.Command("flatpak", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("flatpak repair: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: output}, nil
}

// AddRemote adds a Flatpak remote.
func AddRemote(name string, url string) (Result, error) {
	if name == "" || url == "" {
		return Result{Status: "failed", Error: "name and url are required"}, fmt.Errorf("name and url are required")
	}

	cmd := exec.Command("flatpak", "remote-add", "--if-not-exists", name, url)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("flatpak remote-add: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}, nil
}
