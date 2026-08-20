// Package openvpn provides OpenVPN management operations.
package openvpn

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// StatusResult represents the result of getting OpenVPN status.
type StatusResult struct {
	Running    bool     `json:"running"`
	Version    string   `json:"version,omitempty"`
	ActiveVPN []string `json:"active_vpn,omitempty"`
}

// ClientStatus represents the status of connected clients.
type ClientStatus struct {
	CommonName string `json:"common_name"`
	RealAddr   string `json:"real_address"`
	VirtualAddr string `json:"virtual_address"`
	BytesRecv  string `json:"bytes_received"`
	BytesSent  string `json:"bytes_sent"`
	ConnectedSince string `json:"connected_since"`
}

// ActionResult represents the result of an OpenVPN action.
type ActionResult struct {
	Changed    bool   `json:"changed"`
	Action     string `json:"action"`
	DurationMs int64  `json:"duration_ms"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Status returns the OpenVPN service status.
func Status() (*StatusResult, error) {
	result := &StatusResult{}

	// Check if service is running
	cmd := exec.Command("systemctl", "is-active", "openvpn")
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == "active" {
		result.Running = true
	}

	// Get version
	cmd = exec.Command("openvpn", "--version")
	out, err = cmd.CombinedOutput()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "OpenVPN") {
				result.Version = strings.TrimSpace(line)
				break
			}
		}
	}

	return result, nil
}

// Start starts the OpenVPN service.
func Start() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("systemctl", "start", "openvpn")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "start",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("start openvpn: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "start",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Stop stops the OpenVPN service.
func Stop() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("systemctl", "stop", "openvpn")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "stop",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("stop openvpn: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "stop",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Restart restarts the OpenVPN service.
func Restart() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("systemctl", "restart", "openvpn")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "restart",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("restart openvpn: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "restart",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Enable enables OpenVPN to start at boot.
func Enable() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("systemctl", "enable", "openvpn")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "enable",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("enable openvpn: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "enable",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Disable disables OpenVPN from starting at boot.
func Disable() (*ActionResult, error) {
	start := time.Now()
	cmd := exec.Command("systemctl", "disable", "openvpn")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "disable",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("disable openvpn: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "disable",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// GenKey generates a static key for OpenVPN (tls-auth or secret).
func GenKey(outputPath string) (*ActionResult, error) {
	if outputPath == "" {
		return nil, fmt.Errorf("output path is required")
	}

	start := time.Now()
	cmd := exec.Command("openvpn", "--genkey", "--secret", outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "genkey",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("generate key: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "genkey",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     outputPath,
	}, nil
}

// GenTLSAuth generates a tls-auth key for OpenVPN.
func GenTLSAuth(outputPath string) (*ActionResult, error) {
	if outputPath == "" {
		return nil, fmt.Errorf("output path is required")
	}

	start := time.Now()
	cmd := exec.Command("openvpn", "--genkey", "secret", outputPath)
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try legacy syntax
		cmd = exec.Command("openvpn", "--genkey", "--secret", outputPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return &ActionResult{
				Changed:    false,
				Action:     "gen_tls_auth",
				DurationMs: time.Since(start).Milliseconds(),
				Output:     string(output),
				Error:      err.Error(),
			}, fmt.Errorf("generate tls-auth key: %s", string(output))
		}
	}
	return &ActionResult{
		Changed:    true,
		Action:     "gen_tls_auth",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     outputPath,
	}, nil
}
