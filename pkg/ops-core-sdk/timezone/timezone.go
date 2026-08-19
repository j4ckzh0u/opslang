// Package timezone provides timezone management operations.
package timezone

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TimezoneResult represents the result of getting timezone.
type TimezoneResult struct {
	Timezone string `json:"timezone"`
}

// ListResult represents the result of listing timezones.
type ListResult struct {
	Zones []string `json:"zones"`
}

// ActionResult represents the result of setting timezone.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// Get returns the current timezone.
func Get() (*TimezoneResult, error) {
	// Try timedatectl first (systemd systems)
	if out, err := exec.Command("timedatectl", "show", "--property=Timezone", "--value").CombinedOutput(); err == nil {
		tz := strings.TrimSpace(string(out))
		if tz != "" {
			return &TimezoneResult{Timezone: tz}, nil
		}
	}

	// Fallback to reading /etc/timezone
	if data, err := os.ReadFile("/etc/timezone"); err == nil {
		tz := strings.TrimSpace(string(data))
		if tz != "" {
			return &TimezoneResult{Timezone: tz}, nil
		}
	}

	// Fallback to /etc/localtime symlink
	if target, err := os.Readlink("/etc/localtime"); err == nil {
		// Extract timezone from path like /usr/share/zoneinfo/America/New_York
		if idx := strings.Index(target, "zoneinfo/"); idx != -1 {
			tz := target[idx+9:]
			return &TimezoneResult{Timezone: tz}, nil
		}
	}

	return &TimezoneResult{Timezone: "UTC"}, nil
}

// Set sets the system timezone.
func Set(timezone string) (*ActionResult, error) {
	// Validate timezone
	zoneinfoPath := fmt.Sprintf("/usr/share/zoneinfo/%s", timezone)
	if _, err := os.Stat(zoneinfoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("timezone %s does not exist", timezone)
	}

	// Try timedatectl first (systemd systems)
	if _, err := exec.LookPath("timedatectl"); err == nil {
		cmd := exec.Command("timedatectl", "set-timezone", timezone)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("timedatectl set-timezone failed: %w", err)
		}
		return &ActionResult{
			Changed: true,
			Message: fmt.Sprintf("Timezone set to %s (via timedatectl)", timezone),
		}, nil
	}

	// Fallback to manual symlink
	if err := os.Remove("/etc/localtime"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove /etc/localtime: %w", err)
	}

	if err := os.Symlink(zoneinfoPath, "/etc/localtime"); err != nil {
		return nil, fmt.Errorf("failed to create symlink: %w", err)
	}

	// Update /etc/timezone
	if err := os.WriteFile("/etc/timezone", []byte(timezone+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to write /etc/timezone: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Timezone set to %s", timezone),
	}, nil
}

// List returns all available timezones.
func List() (*ListResult, error) {
	out, err := exec.Command("timedatectl", "list-timezones").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("timedatectl list-timezones failed: %w", err)
	}

	zones := strings.Split(strings.TrimSpace(string(out)), "\n")
	return &ListResult{Zones: zones}, nil
}
