// Package hostname provides hostname management operations.
package hostname

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HostnameResult represents the result of getting hostname.
type HostnameResult struct {
	Hostname string `json:"hostname"`
	FQDN     string `json:"fqdn"`
}

// ActionResult represents the result of setting hostname.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// Get returns the current hostname.
func Get() (*HostnameResult, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Try to get FQDN
	fqdn := hostname
	if out, err := exec.Command("hostname", "-f").CombinedOutput(); err == nil {
		fqdn = strings.TrimSpace(string(out))
	}

	return &HostnameResult{
		Hostname: hostname,
		FQDN:     fqdn,
	}, nil
}

// Set sets the system hostname.
func Set(hostname string) (*ActionResult, error) {
	// Set hostname using hostnamectl (systemd systems)
	if _, err := exec.LookPath("hostnamectl"); err == nil {
		cmd := exec.Command("hostnamectl", "set-hostname", hostname)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("hostnamectl set-hostname failed: %w", err)
		}
		return &ActionResult{
			Changed: true,
			Message: fmt.Sprintf("Hostname set to %s (via hostnamectl)", hostname),
		}, nil
	}

	// Fallback to hostname command
	cmd := exec.Command("hostname", hostname)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("hostname command failed: %w", err)
	}

	// Update /etc/hostname
	if err := os.WriteFile("/etc/hostname", []byte(hostname+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to write /etc/hostname: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Hostname set to %s", hostname),
	}, nil
}

// SetFQDN sets the fully qualified domain name.
func SetFQDN(fqdn string) (*ActionResult, error) {
	// Extract short hostname from FQDN
	parts := strings.SplitN(fqdn, ".", 2)
	hostname := parts[0]

	// Set hostname first
	if _, err := Set(hostname); err != nil {
		return nil, err
	}

	// Update /etc/hosts
	hostsFile := "/etc/hosts"
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/hosts: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	updated := false
	for i, line := range lines {
		if strings.HasPrefix(line, "127.0.1.1") || strings.HasPrefix(line, "127.0.0.1") {
			// Update the line
			fields := strings.Fields(line)
			if len(fields) > 0 {
				lines[i] = fmt.Sprintf("%s\t%s\t%s", fields[0], fqdn, hostname)
				updated = true
				break
			}
		}
	}

	if !updated {
		// Add new entry
		lines = append(lines, fmt.Sprintf("127.0.1.1\t%s\t%s", fqdn, hostname))
	}

	if err := os.WriteFile(hostsFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return nil, fmt.Errorf("failed to write /etc/hosts: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("FQDN set to %s", fqdn),
	}, nil
}
