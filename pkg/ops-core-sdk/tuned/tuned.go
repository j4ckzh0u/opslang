// Package tuned provides system tuning profile management.
package tuned

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by tuning operations.
type Result struct {
	Profile string `json:"profile,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// StatusResult is returned by status queries.
type StatusResult struct {
	Active  bool   `json:"active"`
	Profile string `json:"profile,omitempty"`
	Details string `json:"details,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ProfileInfo represents a tuned profile.
type ProfileInfo struct {
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Default bool   `json:"default"`
}

// ProfilesResult is returned by profile listing.
type ProfilesResult struct {
	Profiles []ProfileInfo `json:"profiles"`
	Count    int           `json:"count"`
	Error    string        `json:"error,omitempty"`
}

func tunedAdm(args ...string) (string, error) {
	cmd := exec.Command("tuned-adm", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Set sets the active tuning profile.
func Set(profile string) Result {
	if profile == "" {
		return Result{Error: "profile is required"}
	}
	out, err := tunedAdm("profile", profile)
	if err != nil {
		return Result{Profile: profile, Error: fmt.Sprintf("tuned-adm profile failed: %s: %s", err, out)}
	}
	return Result{Profile: profile, Success: true, Changed: true}
}

// Status returns current tuning status.
func Status() StatusResult {
	out, err := tunedAdm("active")
	if err != nil {
		return StatusResult{Error: fmt.Sprintf("tuned-adm active failed: %s: %s", err, out)}
	}
	result := StatusResult{Active: true}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Current active profile") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result.Profile = strings.TrimSpace(parts[1])
			}
		}
	}
	return result
}

// List lists available profiles.
func List() ProfilesResult {
	out, err := tunedAdm("list")
	if err != nil {
		return ProfilesResult{Error: fmt.Sprintf("tuned-adm list failed: %s: %s", err, out)}
	}
	var profiles []ProfileInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Available") || strings.HasPrefix(line, "Current") {
			continue
		}
		// Lines like "- throughput-performance" or "* throughput-performance"
		prefix := ""
		if strings.HasPrefix(line, "- ") {
			line = strings.TrimPrefix(line, "- ")
		} else if strings.HasPrefix(line, "* ") {
			line = strings.TrimPrefix(line, "* ")
			prefix = "active"
		}
		if line != "" {
			profiles = append(profiles, ProfileInfo{
				Name:   line,
				Active: prefix == "active",
			})
		}
	}
	return ProfilesResult{Profiles: profiles, Count: len(profiles)}
}

// Off disables tuned service.
func Off() Result {
	out, err := tunedAdm("off")
	if err != nil {
		return Result{Error: fmt.Sprintf("tuned-adm off failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true}
}

// Profile returns the name of the current active profile.
func Profile() (string, error) {
	out, err := tunedAdm("active")
	if err != nil {
		return "", fmt.Errorf("tuned-adm active failed: %w: %s", err, out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Current active profile") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}
	return "", nil
}

// Verify verifies the current tuning.
func Verify() Result {
	out, err := tunedAdm("verify")
	if err != nil {
		return Result{Error: fmt.Sprintf("tuned-adm verify failed: %s: %s", err, out)}
	}
	return Result{Success: true, Output: strings.TrimSpace(out)}
}
