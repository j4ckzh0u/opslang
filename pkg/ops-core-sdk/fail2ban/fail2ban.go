// Package fail2ban provides fail2ban jail management operations.
package fail2ban

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// StatusResult represents the result of getting fail2ban status.
type StatusResult struct {
	Running   bool     `json:"running"`
	JailCount int      `json:"jail_count"`
	Jails     []string `json:"jails"`
}

// JailStatusResult represents the status of a specific jail.
type JailStatusResult struct {
	Jail       string   `json:"jail"`
	Enabled    bool     `json:"enabled"`
	Failed     int      `json:"failed"`
	Banned     int      `json:"banned"`
	IPs        []string `json:"banned_ips,omitempty"`
}

// ActionResult represents the result of a fail2ban action.
type ActionResult struct {
	Changed    bool   `json:"changed"`
	Action     string `json:"action"`
	Jail       string `json:"jail,omitempty"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// Get returns the overall fail2ban status.
func Get() (*StatusResult, error) {
	cmd := exec.Command("fail2ban-client", "status")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &StatusResult{Running: false}, nil
	}

	result := &StatusResult{Running: true}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Number of jail:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &result.JailCount)
			}
		}
		if strings.HasPrefix(line, "Jail list:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				jails := strings.TrimSpace(parts[1])
				if jails != "" {
					for _, j := range strings.Split(jails, ",") {
						result.Jails = append(result.Jails, strings.TrimSpace(j))
					}
				}
			}
		}
	}
	if result.JailCount == 0 {
		result.JailCount = len(result.Jails)
	}

	return result, nil
}

// JailStatus returns the status of a specific jail.
func JailStatus(jail string) (*JailStatusResult, error) {
	if jail == "" {
		return nil, fmt.Errorf("jail name is required")
	}

	cmd := exec.Command("fail2ban-client", "status", jail)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("fail2ban-client status %s: %s", jail, string(out))
	}

	result := &JailStatusResult{Jail: jail, Enabled: true}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Currently failed:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &result.Failed)
			}
		}
		if strings.Contains(line, "Currently banned:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &result.Banned)
			}
		}
		if strings.Contains(line, "Banned IP list:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ips := strings.TrimSpace(parts[1])
				if ips != "" {
					for _, ip := range strings.Fields(ips) {
						result.IPs = append(result.IPs, ip)
					}
				}
			}
		}
	}

	return result, nil
}

// Start starts the fail2ban service.
func Start() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("systemctl", "start", "fail2ban")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "start",
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("start fail2ban: %s", output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "start",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Stop stops the fail2ban service.
func Stop() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("systemctl", "stop", "fail2ban")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "stop",
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("stop fail2ban: %s", output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "stop",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Reload reloads fail2ban configuration.
func Reload() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("fail2ban-client", "reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "reload",
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("reload fail2ban: %s", output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "reload",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// BanIP bans an IP in a specific jail.
func BanIP(jail, ip string) (*ActionResult, error) {
	if jail == "" {
		return nil, fmt.Errorf("jail name is required")
	}
	if ip == "" {
		return nil, fmt.Errorf("IP address is required")
	}

	start := time.Now()
	cmd := exec.Command("fail2ban-client", "set", jail, "banip", ip)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "ban",
			Jail:       jail,
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("ban IP %s in jail %s: %s", ip, jail, output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "ban",
		Jail:       jail,
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// UnbanIP unbans an IP from a specific jail.
func UnbanIP(jail, ip string) (*ActionResult, error) {
	if jail == "" {
		return nil, fmt.Errorf("jail name is required")
	}
	if ip == "" {
		return nil, fmt.Errorf("IP address is required")
	}

	start := time.Now()
	cmd := exec.Command("fail2ban-client", "set", jail, "unbanip", ip)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "unban",
			Jail:       jail,
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("unban IP %s from jail %s: %s", ip, jail, output)
	}
	return &ActionResult{
		Changed:    true,
		Action:     "unban",
		Jail:       jail,
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}
