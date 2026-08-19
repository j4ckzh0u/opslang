// Package gem provides Ruby gem package management.
// Uses exec.Command to invoke gem binary.
package gem

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ActionResult is returned by Install/Uninstall/Update.
type ActionResult struct {
	Name     string `json:"name"`
	Success  bool   `json:"success"`
	Changed  bool   `json:"changed"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// GemInfo represents a gem's metadata.
type GemInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Default  bool   `json:"default"`
}

// InfoResult is returned by Info.
type InfoResult struct {
	Name  string  `json:"name"`
	Found bool    `json:"found"`
	Info  GemInfo `json:"info,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Gems  []GemInfo `json:"gems"`
	Count int       `json:"count"`
}

// VersionResult is returned by Version.
type VersionResult struct {
	Version string `json:"version"`
	Raw     string `json:"raw"`
}

// Install installs a gem.
func Install(name, version string, userInstall bool) (ActionResult, error) {
	start := time.Now()
	if name == "" {
		return ActionResult{}, fmt.Errorf("gem name is required")
	}
	args := []string{"install", name, "--no-document"}
	if version != "" {
		args = append(args, "-v", version)
	}
	if userInstall {
		args = append(args, "--user-install")
	}
	out, err := exec.Command("gem", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Error: string(out), Duration: time.Since(start).Milliseconds()}, err
	}
	changed := !strings.Contains(string(out), "already installed")
	return ActionResult{Name: name, Success: true, Changed: changed, Duration: time.Since(start).Milliseconds()}, nil
}

// Uninstall uninstalls a gem.
func Uninstall(name string, force bool) (ActionResult, error) {
	start := time.Now()
	if name == "" {
		return ActionResult{}, fmt.Errorf("gem name is required")
	}
	args := []string{"uninstall", name, "-x"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, "--no-document")
	out, err := exec.Command("gem", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Error: string(out), Duration: time.Since(start).Milliseconds()}, err
	}
	return ActionResult{Name: name, Success: true, Changed: true, Duration: time.Since(start).Milliseconds()}, nil
}

// Update updates a gem (or all gems if name is empty).
func Update(name string) (ActionResult, error) {
	start := time.Now()
	args := []string{"update"}
	if name != "" {
		args = append(args, name)
	}
	args = append(args, "--no-document")
	out, err := exec.Command("gem", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Error: string(out), Duration: time.Since(start).Milliseconds()}, err
	}
	return ActionResult{Name: name, Success: true, Changed: true, Duration: time.Since(start).Milliseconds()}, nil
}

// Info returns metadata about a gem.
func Info(name string) (InfoResult, error) {
	if name == "" {
		return InfoResult{}, fmt.Errorf("gem name is required")
	}
	out, err := exec.Command("gem", "list", "^"+name+"$", "--exact").Output()
	if err != nil {
		return InfoResult{Name: name, Found: false}, nil
	}
	output := strings.TrimSpace(string(out))
	if output == "" || strings.Contains(output, "*** LOCAL GEMS ***") {
		// Check if there's actually a gem line
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, name) {
				return parseGemLine(line, name)
			}
		}
		return InfoResult{Name: name, Found: false}, nil
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, name) {
			return parseGemLine(line, name)
		}
	}
	return InfoResult{Name: name, Found: false}, nil
}

func parseGemLine(line, name string) (InfoResult, error) {
	// Format: "gemname (version, version2, ...)"
	start := strings.Index(line, "(")
	end := strings.Index(line, ")")
	if start < 0 || end < 0 || end <= start {
		return InfoResult{Name: name, Found: true, Info: GemInfo{Name: name}}, nil
	}
	versions := strings.Split(line[start+1:end], ",")
	version := strings.TrimSpace(versions[0])
	return InfoResult{Name: name, Found: true, Info: GemInfo{Name: name, Version: version}}, nil
}

// List returns all installed gems.
func List() (ListResult, error) {
	out, err := exec.Command("gem", "list", "--local").Output()
	if err != nil {
		return ListResult{}, fmt.Errorf("failed to list gems: %w", err)
	}
	var gems []GemInfo
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "***") {
			continue
		}
		start := strings.Index(line, "(")
		end := strings.Index(line, ")")
		if start < 0 || end < 0 || end <= start {
			continue
		}
		name := strings.TrimSpace(line[:start])
		versions := strings.Split(line[start+1:end], ",")
		version := strings.TrimSpace(versions[0])
		gems = append(gems, GemInfo{Name: name, Version: version})
	}
	return ListResult{Gems: gems, Count: len(gems)}, nil
}

// Version returns the gem command version.
func Version() (VersionResult, error) {
	out, err := exec.Command("gem", "--version").CombinedOutput()
	if err != nil {
		return VersionResult{}, fmt.Errorf("failed to get gem version: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	return VersionResult{Version: raw, Raw: raw}, nil
}
