// Package snap provides Snap package management operations.
package snap

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a snap operation.
type ActionResult struct {
	Name    string `json:"name"`
	Channel string `json:"channel,omitempty"`
	Changed bool   `json:"changed"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// SnapInfo represents information about a snap package.
type SnapInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Rev         string `json:"rev"`
	Tracking    string `json:"tracking"`
	Publisher   string `json:"publisher"`
	Contact     string `json:"contact,omitempty"`
	Summary     string `json:"summary"`
	Notes       string `json:"notes"`
}

// ListResult represents the result of listing snaps.
type ListResult struct {
	Snaps []SnapInfo `json:"snaps"`
}

// Install installs a snap package.
func Install(name string, channel string, classic bool) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("snap name is required")
	}

	args := []string{"install", "--yes"}
	if classic {
		args = append(args, "--classic")
	}
	if channel != "" {
		args = append(args, "--channel", channel)
	}
	args = append(args, name)

	out, err := exec.Command("snap", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("snap install failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Channel: channel, Changed: true, Success: true}, nil
}

// Remove removes a snap package.
func Remove(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("snap name is required")
	}

	out, err := exec.Command("snap", "remove", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("snap remove failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Refresh refreshes a snap package.
func Refresh(name string, channel string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("snap name is required")
	}

	args := []string{"refresh"}
	if channel != "" {
		args = append(args, "--channel", channel)
	}
	args = append(args, name)

	out, err := exec.Command("snap", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("snap refresh failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Channel: channel, Changed: true, Success: true}, nil
}

// List lists installed snaps.
func List() (ListResult, error) {
	out, err := exec.Command("snap", "list", "--unicode=never").CombinedOutput()
	if err != nil {
		return ListResult{}, fmt.Errorf("snap list failed: %w (output: %s)", err, string(out))
	}

	result := ListResult{Snaps: make([]SnapInfo, 0)}
	lines := strings.Split(string(out), "\n")
	// Skip header line
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		// Parse: Name Version Rev Tracking Publisher Notes
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			snap := SnapInfo{
				Name:    fields[0],
				Version: fields[1],
				Rev:     fields[2],
			}
			if len(fields) >= 4 {
				snap.Tracking = fields[3]
			}
			if len(fields) >= 5 {
				snap.Publisher = fields[4]
			}
			if len(fields) >= 7 {
				snap.Summary = strings.Join(fields[5:len(fields)-1], " ")
				snap.Notes = fields[len(fields)-1]
			}
			result.Snaps = append(result.Snaps, snap)
		}
	}
	return result, nil
}

// Get gets information about a specific snap.
func Get(name string) (SnapInfo, error) {
	if name == "" {
		return SnapInfo{}, fmt.Errorf("snap name is required")
	}

	// snap info outputs text, we need to parse it
	out, err := exec.Command("snap", "info", name).CombinedOutput()
	if err != nil {
		return SnapInfo{}, fmt.Errorf("snap info failed: %w (output: %s)", err, string(out))
	}

	snap := SnapInfo{Name: name}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			snap.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		} else if strings.HasPrefix(line, "version:") {
			snap.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		} else if strings.HasPrefix(line, "rev:") {
			snap.Rev = strings.TrimSpace(strings.TrimPrefix(line, "rev:"))
		} else if strings.HasPrefix(line, "tracking:") {
			snap.Tracking = strings.TrimSpace(strings.TrimPrefix(line, "tracking:"))
		} else if strings.HasPrefix(line, "publisher:") {
			snap.Publisher = strings.TrimSpace(strings.TrimPrefix(line, "publisher:"))
		} else if strings.HasPrefix(line, "contact:") {
			snap.Contact = strings.TrimSpace(strings.TrimPrefix(line, "contact:"))
		} else if strings.HasPrefix(line, "summary:") {
			snap.Summary = strings.TrimSpace(strings.TrimPrefix(line, "summary:"))
		}
	}
	return snap, nil
}

// Enable enables a snap.
func Enable(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("snap name is required")
	}

	out, err := exec.Command("snap", "enable", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("snap enable failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Disable disables a snap.
func Disable(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("snap name is required")
	}

	out, err := exec.Command("snap", "disable", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("snap disable failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Switch switches a snap to a different channel.
func Switch(name string, channel string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("snap name is required")
	}
	if channel == "" {
		return ActionResult{}, fmt.Errorf("channel is required")
	}

	out, err := exec.Command("snap", "switch", "--channel", channel, name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("snap switch failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Channel: channel, Changed: true, Success: true}, nil
}

// Changes lists snap changes.
type ChangeInfo struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

// ChangesResult represents the result of listing changes.
type ChangesResult struct {
	Changes []ChangeInfo `json:"changes"`
}

// Changes lists recent snap changes.
func Changes() (ChangesResult, error) {
	out, err := exec.Command("snap", "changes").CombinedOutput()
	if err != nil {
		return ChangesResult{}, fmt.Errorf("snap changes failed: %w (output: %s)", err, string(out))
	}

	result := ChangesResult{Changes: make([]ChangeInfo, 0)}
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		// Parse: Id   Status   Spawn                Ready                Summary
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			change := ChangeInfo{
				ID:      fields[0],
				Status:  fields[1],
				Summary: strings.Join(fields[4:], " "),
			}
			result.Changes = append(result.Changes, change)
		}
	}
	return result, nil
}

// JSON output helper
func (r ListResult) JSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
