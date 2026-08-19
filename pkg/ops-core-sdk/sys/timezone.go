// Timezone operations.
package sys

import (
	"fmt"
	"os"
	"strings"
)

// TimezoneResult is returned by TimezoneGet.
type TimezoneResult struct {
	Timezone string `json:"timezone"`
}

// TimezoneSetResult is returned by TimezoneSet.
type TimezoneSetResult struct {
	Timezone string `json:"timezone"`
	Changed  bool   `json:"changed"`
}

// TimezoneGet returns the current system timezone.
func TimezoneGet() (TimezoneResult, error) {
	// Try timedatectl first (systemd)
	// Fall back to reading /etc/timezone or /etc/localtime
	data, err := os.ReadFile("/etc/timezone")
	if err == nil {
		tz := strings.TrimSpace(string(data))
		if tz != "" {
			return TimezoneResult{Timezone: tz}, nil
		}
	}

	// Fall back to /etc/localtime symlink
	target, err := os.Readlink("/etc/localtime")
	if err == nil {
		// Extract timezone from path like /usr/share/zoneinfo/America/New_York
		parts := strings.Split(target, "zoneinfo/")
		if len(parts) == 2 {
			return TimezoneResult{Timezone: parts[1]}, nil
		}
	}

	return TimezoneResult{Timezone: "UTC"}, nil
}

// TimezoneSet sets the system timezone.
func TimezoneSet(timezone string) (TimezoneSetResult, error) {
	result := TimezoneSetResult{Timezone: timezone}

	// Check current timezone
	current, err := TimezoneGet()
	if err == nil && current.Timezone == timezone {
		result.Changed = false
		return result, nil
	}

	// Verify timezone file exists
	zoneinfoPath := "/usr/share/zoneinfo/" + timezone
	if _, err := os.Stat(zoneinfoPath); err != nil {
		return result, fmt.Errorf("sys.TimezoneSet: timezone %q not found: %w", timezone, err)
	}

	// Write /etc/timezone
	if err := os.WriteFile("/etc/timezone", []byte(timezone+"\n"), 0644); err != nil {
		return result, fmt.Errorf("sys.TimezoneSet: %w", err)
	}

	// Symlink /etc/localtime
	os.Remove("/etc/localtime")
	if err := os.Symlink(zoneinfoPath, "/etc/localtime"); err != nil {
		return result, fmt.Errorf("sys.TimezoneSet: %w", err)
	}

	result.Changed = true
	return result, nil
}
