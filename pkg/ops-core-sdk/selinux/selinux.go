// Package selinux provides SELinux status and configuration management.
package selinux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Status represents SELinux status.
type Status struct {
	Enabled   bool   `json:"enabled"`
	Mode      string `json:"mode"`      // enforcing, permissive, disabled
	Policy    string `json:"policy"`    // targeted, mls, etc.
	ConfigFile string `json:"config_file"`
}

// GetResult is returned by Get.
type GetResult struct {
	Status Status `json:"status"`
}

// SetResult is returned by Set.
type SetResult struct {
	Changed bool   `json:"changed"`
	Mode    string `json:"mode"`
	Error   string `json:"error,omitempty"`
}

const selinuxConfig = "/etc/selinux/config"

// Get returns the current SELinux status.
func Get() (GetResult, error) {
	status := Status{ConfigFile: selinuxConfig}

	// Try getenforce first
	out, err := exec.Command("getenforce").Output()
	if err == nil {
		mode := strings.TrimSpace(strings.ToLower(string(out)))
		status.Mode = mode
		status.Enabled = mode != "disabled"
	}

	// Try to get policy from config
	data, err := os.ReadFile(selinuxConfig)
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "SELINUXTYPE=") {
				status.Policy = strings.TrimPrefix(line, "SELINUXTYPE=")
			}
		}
	}

	return GetResult{Status: status}, nil
}

// Set changes the SELinux mode (enforcing/permissive).
func Set(mode string) (SetResult, error) {
	if mode == "" {
		return SetResult{Error: "mode is required"}, fmt.Errorf("mode is required")
	}

	mode = strings.ToLower(mode)
	if mode != "enforcing" && mode != "permissive" && mode != "disabled" {
		return SetResult{Error: "invalid mode"}, fmt.Errorf("mode must be enforcing, permissive, or disabled")
	}

	// Check current mode
	current, _ := Get()
	if strings.ToLower(current.Status.Mode) == mode {
		return SetResult{Changed: false, Mode: mode}, nil
	}

	if mode == "disabled" {
		// Can only be disabled via config + reboot
		if err := updateConfigFile("SELINUX=disabled"); err != nil {
			return SetResult{Error: err.Error()}, err
		}
		return SetResult{Changed: true, Mode: mode}, nil
	}

	// Use setenforce for runtime change
	var cmd *exec.Cmd
	if mode == "enforcing" {
		cmd = exec.Command("setenforce", "1")
	} else {
		cmd = exec.Command("setenforce", "0")
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return SetResult{Error: string(output)}, fmt.Errorf("setenforce: %s", output)
	}

	// Update config file for persistence
	selinuxVal := "SELINUX=" + mode
	if err := updateConfigFile(selinuxVal); err != nil {
		return SetResult{Error: err.Error()}, err
	}

	return SetResult{Changed: true, Mode: mode}, nil
}

// updateConfigFile updates the SELINUX= line in the config file.
func updateConfigFile(newLine string) error {
	data, err := os.ReadFile(selinuxConfig)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var lines []string
	found := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "SELINUX=") && !strings.HasPrefix(trimmed, "#") {
			lines = append(lines, newLine)
			found = true
		} else {
			lines = append(lines, line)
		}
	}

	if !found {
		lines = append(lines, newLine)
	}

	return os.WriteFile(selinuxConfig, []byte(strings.Join(lines, "\n")), 0644)
}
