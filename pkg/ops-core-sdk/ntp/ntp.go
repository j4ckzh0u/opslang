// Package ntp provides NTP/chrony time synchronization management.
package ntp

import (
	"fmt"
	"os/exec"
	"strings"
)

// Status represents NTP synchronization status.
type Status struct {
	Synchronized bool   `json:"synchronized"`
	Server       string `json:"server"`
	Offset       string `json:"offset"`
	Daemon       string `json:"daemon"` // chronyd, ntpd, systemd-timesyncd
}

// GetResult is returned by Get.
type GetResult struct {
	Status Status `json:"status"`
}

// SetResult is returned by Set.
type SetResult struct {
	Changed bool   `json:"changed"`
	Server  string `json:"server"`
	Error   string `json:"error,omitempty"`
}

// Get returns the current NTP synchronization status.
func Get() (GetResult, error) {
	status := Status{}

	// Try chronyc first (chrony)
	out, err := exec.Command("chronyc", "tracking").Output()
	if err == nil {
		status.Daemon = "chronyd"
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "Reference ID") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					status.Server = strings.TrimSpace(strings.Split(parts[1], "(")[0])
				}
			}
			if strings.HasPrefix(line, "System time") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					status.Offset = strings.TrimSpace(strings.Split(parts[1], " ")[0])
				}
			}
		}
		status.Synchronized = status.Server != "" && status.Server != "()"
		return GetResult{Status: status}, nil
	}

	// Try ntpq (ntpd)
	out, err = exec.Command("ntpq", "-p").Output()
	if err == nil {
		status.Daemon = "ntpd"
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "*") {
				fields := strings.Fields(line)
				if len(fields) > 0 {
					status.Server = strings.TrimPrefix(fields[0], "*")
					status.Synchronized = true
				}
				break
			}
		}
		return GetResult{Status: status}, nil
	}

	// Try timedatectl (systemd-timesyncd)
	out, err = exec.Command("timedatectl", "show").Output()
	if err == nil {
		status.Daemon = "systemd-timesyncd"
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "NTPSynchronized=") {
				status.Synchronized = strings.TrimPrefix(line, "NTPSynchronized=") == "yes"
			}
		}
		return GetResult{Status: status}, nil
	}

	return GetResult{}, fmt.Errorf("no NTP daemon detected")
}

// Set configures the NTP server (chrony only).
func Set(server string) (SetResult, error) {
	if server == "" {
		return SetResult{Error: "server is required"}, fmt.Errorf("server is required")
	}

	// Check if chrony is available
	if _, err := exec.LookPath("chronyc"); err != nil {
		return SetResult{Error: "chrony not installed"}, fmt.Errorf("chrony not installed")
	}

	// Add server to chrony config
	cmd := exec.Command("chronyc", "add", "server", server)
	if output, err := cmd.CombinedOutput(); err != nil {
		return SetResult{Error: string(output)}, fmt.Errorf("chronyc add server: %s", output)
	}

	return SetResult{Changed: true, Server: server}, nil
}
