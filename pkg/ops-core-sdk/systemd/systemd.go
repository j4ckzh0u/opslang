// Package systemd provides systemd unit management operations.
// Manages enable/disable, daemon-reload, mask/unmask, active/enabled status checks.
package systemd

import (
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a systemd action.
type ActionResult struct {
	Unit    string `json:"unit"`
	Action  string `json:"action"`
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
}

// StatusResult represents the result of a status check.
type StatusResult struct {
	Unit      string `json:"unit"`
	Active    bool   `json:"active"`
	Enabled   bool   `json:"enabled"`
	Loaded    bool   `json:"loaded"`
	State     string `json:"state"`    // active, inactive, failed, activating
	LoadState string `json:"load_state"` // loaded, not-found, masked, error
	SubState  string `json:"sub_state"`  // running, dead, exited, etc.
}

// UnitInfo represents detailed information about a systemd unit.
type UnitInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LoadState   string `json:"load_state"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
	Fragment    string `json:"fragment_path,omitempty"`
	Enabled     bool   `json:"enabled"`
	Masked      bool   `json:"masked"`
	Static      bool   `json:"static"`
}

// ListResult represents the result of listing units.
type ListResult struct {
	Units []UnitInfo `json:"units"`
	Count int        `json:"count"`
}

// runSystemctl executes a systemctl command and returns output.
func runSystemctl(args ...string) (string, error) {
	cmd := exec.Command("systemctl", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// IsActive checks if a unit is currently active (running).
func IsActive(unit string) (*StatusResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.IsActive: unit name is required")
	}
	result := &StatusResult{Unit: unit}

	state, _ := runSystemctl("is-active", unit)
	result.State = state
	result.Active = (state == "active")

	loadState, _ := runSystemctl("show", "--property=LoadState", "--value", unit)
	result.LoadState = loadState
	result.Loaded = (loadState == "loaded")

	subState, _ := runSystemctl("show", "--property=SubState", "--value", unit)
	result.SubState = subState

	enabled, _ := runSystemctl("is-enabled", unit)
	result.Enabled = (enabled == "enabled")

	return result, nil
}

// IsEnabled checks if a unit is enabled to start at boot.
func IsEnabled(unit string) (*StatusResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.IsEnabled: unit name is required")
	}
	result := &StatusResult{Unit: unit}

	enabled, _ := runSystemctl("is-enabled", unit)
	result.Enabled = (enabled == "enabled")
	result.State = enabled

	state, _ := runSystemctl("is-active", unit)
	result.State = state
	result.Active = (state == "active")

	loadState, _ := runSystemctl("show", "--property=LoadState", "--value", unit)
	result.LoadState = loadState
	result.Loaded = (loadState == "loaded")

	subState, _ := runSystemctl("show", "--property=SubState", "--value", unit)
	result.SubState = subState

	return result, nil
}

// Enable enables a systemd unit to start at boot.
func Enable(unit string) (*ActionResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Enable: unit name is required")
	}
	result := &ActionResult{Unit: unit, Action: "enable"}

	// Check current state
	enabled, _ := runSystemctl("is-enabled", unit)
	if enabled == "enabled" {
		result.Changed = false
		result.Message = "already enabled"
		return result, nil
	}

	out, err := runSystemctl("enable", unit)
	if err != nil {
		return nil, fmt.Errorf("systemd.Enable: failed to enable %s: %w (%s)", unit, err, out)
	}
	result.Changed = true
	result.Message = "enabled"
	return result, nil
}

// Disable disables a systemd unit from starting at boot.
func Disable(unit string) (*ActionResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Disable: unit name is required")
	}
	result := &ActionResult{Unit: unit, Action: "disable"}

	enabled, _ := runSystemctl("is-enabled", unit)
	if enabled != "enabled" {
		result.Changed = false
		result.Message = "already disabled"
		return result, nil
	}

	out, err := runSystemctl("disable", unit)
	if err != nil {
		return nil, fmt.Errorf("systemd.Disable: failed to disable %s: %w (%s)", unit, err, out)
	}
	result.Changed = true
	result.Message = "disabled"
	return result, nil
}

// Start starts a systemd unit.
func Start(unit string) (*ActionResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Start: unit name is required")
	}
	result := &ActionResult{Unit: unit, Action: "start"}

	// Check if already active
	state, _ := runSystemctl("is-active", unit)
	if state == "active" {
		result.Changed = false
		result.Message = "already running"
		return result, nil
	}

	out, err := runSystemctl("start", unit)
	if err != nil {
		return nil, fmt.Errorf("systemd.Start: failed to start %s: %w (%s)", unit, err, out)
	}
	result.Changed = true
	result.Message = "started"
	return result, nil
}

// Stop stops a systemd unit.
func Stop(unit string) (*ActionResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Stop: unit name is required")
	}
	result := &ActionResult{Unit: unit, Action: "stop"}

	state, _ := runSystemctl("is-active", unit)
	if state != "active" {
		result.Changed = false
		result.Message = "already stopped"
		return result, nil
	}

	out, err := runSystemctl("stop", unit)
	if err != nil {
		return nil, fmt.Errorf("systemd.Stop: failed to stop %s: %w (%s)", unit, err, out)
	}
	result.Changed = true
	result.Message = "stopped"
	return result, nil
}

// Restart restarts a systemd unit.
func Restart(unit string) (*ActionResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Restart: unit name is required")
	}
	result := &ActionResult{Unit: unit, Action: "restart"}

	out, err := runSystemctl("restart", unit)
	if err != nil {
		return nil, fmt.Errorf("systemd.Restart: failed to restart %s: %w (%s)", unit, err, out)
	}
	result.Changed = true
	result.Message = "restarted"
	return result, nil
}

// Reload reloads a systemd unit's configuration without restarting.
func Reload(unit string) (*ActionResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Reload: unit name is required")
	}
	result := &ActionResult{Unit: unit, Action: "reload"}

	out, err := runSystemctl("reload", unit)
	if err != nil {
		return nil, fmt.Errorf("systemd.Reload: failed to reload %s: %w (%s)", unit, err, out)
	}
	result.Changed = true
	result.Message = "reloaded"
	return result, nil
}

// DaemonReload reloads the systemd manager configuration.
// Call this after adding/modifying unit files.
func DaemonReload() (*ActionResult, error) {
	result := &ActionResult{Unit: "system", Action: "daemon-reload"}
	out, err := runSystemctl("daemon-reload")
	if err != nil {
		return nil, fmt.Errorf("systemd.DaemonReload: failed: %w (%s)", err, out)
	}
	result.Changed = true
	result.Message = "daemon reloaded"
	return result, nil
}

// Mask masks a systemd unit, preventing it from being started.
func Mask(unit string) (*ActionResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Mask: unit name is required")
	}
	result := &ActionResult{Unit: unit, Action: "mask"}

	enabled, _ := runSystemctl("is-enabled", unit)
	if enabled == "masked" {
		result.Changed = false
		result.Message = "already masked"
		return result, nil
	}

	out, err := runSystemctl("mask", unit)
	if err != nil {
		return nil, fmt.Errorf("systemd.Mask: failed to mask %s: %w (%s)", unit, err, out)
	}
	result.Changed = true
	result.Message = "masked"
	return result, nil
}

// Unmask unmasks a systemd unit, allowing it to be started.
func Unmask(unit string) (*ActionResult, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Unmask: unit name is required")
	}
	result := &ActionResult{Unit: unit, Action: "unmask"}

	enabled, _ := runSystemctl("is-enabled", unit)
	if enabled != "masked" {
		result.Changed = false
		result.Message = "not masked"
		return result, nil
	}

	out, err := runSystemctl("unmask", unit)
	if err != nil {
		return nil, fmt.Errorf("systemd.Unmask: failed to unmask %s: %w (%s)", unit, err, out)
	}
	result.Changed = true
	result.Message = "unmasked"
	return result, nil
}

// Show returns detailed information about a systemd unit.
func Show(unit string) (*UnitInfo, error) {
	if unit == "" {
		return nil, fmt.Errorf("systemd.Show: unit name is required")
	}
	info := &UnitInfo{Name: unit}

	props := []string{
		"Description", "LoadState", "ActiveState", "SubState",
		"FragmentPath", "UnitFileState",
	}
	for _, prop := range props {
		val, _ := runSystemctl("show", "--property="+prop, "--value", unit)
		switch prop {
		case "Description":
			info.Description = val
		case "LoadState":
			info.LoadState = val
			info.Masked = (val == "masked")
		case "ActiveState":
			info.ActiveState = val
		case "SubState":
			info.SubState = val
		case "FragmentPath":
			info.Fragment = val
		case "UnitFileState":
			info.Enabled = (val == "enabled")
			info.Static = (val == "static")
		}
	}
	return info, nil
}

// List lists all loaded systemd units, optionally filtered by type.
// unitType can be: service, socket, target, timer, etc. Empty string lists all.
func List(unitType string) (*ListResult, error) {
	args := []string{"list-units", "--all", "--no-pager", "--no-legend"}
	if unitType != "" {
		args = append(args, unitType+".*")
	}
	out, _ := runSystemctl(args...)
	result := &ListResult{}

	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		name := strings.TrimSpace(fields[0])
		if name == "" {
			continue
		}
		info := UnitInfo{
			Name:        name,
			LoadState:   fields[1],
			ActiveState: fields[2],
			SubState:    fields[3],
		}
		if len(fields) > 4 {
			info.Description = strings.Join(fields[4:], " ")
		}
		result.Units = append(result.Units, info)
	}
	result.Count = len(result.Units)
	return result, nil
}
