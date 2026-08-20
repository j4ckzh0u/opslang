// Package wireguard provides WireGuard VPN management operations.
package wireguard

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// InterfaceInfo represents a WireGuard interface status.
type InterfaceInfo struct {
	Name      string   `json:"name"`
	PublicKey string   `json:"public_key"`
	ListenPort int     `json:"listen_port"`
	Addresses []string `json:"addresses,omitempty"`
	Peers     []Peer   `json:"peers,omitempty"`
}

// Peer represents a WireGuard peer.
type Peer struct {
	PublicKey           string   `json:"public_key"`
	Endpoint            string   `json:"endpoint,omitempty"`
	AllowedIPs          []string `json:"allowed_ips,omitempty"`
	LatestHandshake     string   `json:"latest_handshake,omitempty"`
	TransferRx          string   `json:"transfer_rx,omitempty"`
	TransferTx          string   `json:"transfer_tx,omitempty"`
}

// ActionResult represents the result of a wireguard action.
type ActionResult struct {
	Interface  string `json:"interface,omitempty"`
	Changed    bool   `json:"changed"`
	Action     string `json:"action"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// Show returns the status of all WireGuard interfaces.
func Show() ([]InterfaceInfo, error) {
	cmd := exec.Command("wg", "show", "all", "dump")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// No interfaces or wg not installed
		return []InterfaceInfo{}, nil
	}

	ifaces := make(map[string]*InterfaceInfo)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ifName := fields[0]
		if _, ok := ifaces[ifName]; !ok {
			ifaces[ifName] = &InterfaceInfo{Name: ifName}
		}
		iface := ifaces[ifName]

		if len(fields) == 5 {
			// Interface line: name privateKey publicKey listenPort fwmark
			iface.PublicKey = fields[2]
			fmt.Sscanf(fields[3], "%d", &iface.ListenPort)
		} else if len(fields) >= 5 {
			// Peer line: name publicKey endpoint allowedIps latestHandshake transferRx transferTx persistentKeepalive
			peer := Peer{PublicKey: fields[1]}
			if fields[2] != "(none)" {
				peer.Endpoint = fields[2]
			}
			if len(fields) > 3 {
				for _, ip := range strings.Split(fields[3], ",") {
					peer.AllowedIPs = append(peer.AllowedIPs, ip)
				}
			}
			if len(fields) > 5 {
				peer.LatestHandshake = fields[5]
			}
			if len(fields) > 7 {
				peer.TransferRx = fields[6]
				peer.TransferTx = fields[7]
			}
			iface.Peers = append(iface.Peers, peer)
		}
	}

	result := make([]InterfaceInfo, 0, len(ifaces))
	for _, iface := range ifaces {
		result = append(result, *iface)
	}
	return result, nil
}

// Up brings up a WireGuard interface.
func Up(interfaceName, configPath string) (*ActionResult, error) {
	if interfaceName == "" {
		return nil, fmt.Errorf("interface name is required")
	}

	start := time.Now()
	args := []string{"up", interfaceName}
	if configPath != "" {
		args = append(args, configPath)
	}
	cmd := exec.Command("wg-quick", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Interface:  interfaceName,
			Changed:    false,
			Action:     "up",
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("wg-quick up %s: %s", interfaceName, output)
	}
	return &ActionResult{
		Interface:  interfaceName,
		Changed:    true,
		Action:     "up",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Down brings down a WireGuard interface.
func Down(interfaceName string) (*ActionResult, error) {
	if interfaceName == "" {
		return nil, fmt.Errorf("interface name is required")
	}

	start := time.Now()
	cmd := exec.Command("wg-quick", "down", interfaceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Interface:  interfaceName,
			Changed:    false,
			Action:     "down",
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("wg-quick down %s: %s", interfaceName, output)
	}
	return &ActionResult{
		Interface:  interfaceName,
		Changed:    true,
		Action:     "down",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// AddPeer adds a peer to a WireGuard interface.
func AddPeer(interfaceName, publicKey, allowedIPs, endpoint string) (*ActionResult, error) {
	if interfaceName == "" {
		return nil, fmt.Errorf("interface name is required")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("public key is required")
	}

	start := time.Now()
	args := []string{"set", interfaceName, "peer", publicKey}
	if allowedIPs != "" {
		args = append(args, "allowed-ips", allowedIPs)
	}
	if endpoint != "" {
		args = append(args, "endpoint", endpoint)
	}
	cmd := exec.Command("wg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Interface:  interfaceName,
			Changed:    false,
			Action:     "add_peer",
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("add peer: %s", output)
	}
	return &ActionResult{
		Interface:  interfaceName,
		Changed:    true,
		Action:     "add_peer",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// RemovePeer removes a peer from a WireGuard interface.
func RemovePeer(interfaceName, publicKey string) (*ActionResult, error) {
	if interfaceName == "" {
		return nil, fmt.Errorf("interface name is required")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("public key is required")
	}

	start := time.Now()
	cmd := exec.Command("wg", "set", interfaceName, "peer", publicKey, "remove")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Interface:  interfaceName,
			Changed:    false,
			Action:     "remove_peer",
			DurationMs: time.Since(start).Milliseconds(),
			Error:      string(output),
		}, fmt.Errorf("remove peer: %s", output)
	}
	return &ActionResult{
		Interface:  interfaceName,
		Changed:    true,
		Action:     "remove_peer",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// GenKey generates a new WireGuard private key.
func GenKey() (string, error) {
	cmd := exec.Command("wg", "genkey")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("wg genkey: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GenPSK generates a new WireGuard preshared key.
func GenPSK() (string, error) {
	cmd := exec.Command("wg", "genpsk")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("wg genpsk: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// PubKey derives the public key from a private key.
func PubKey(privateKey string) (string, error) {
	if privateKey == "" {
		return "", fmt.Errorf("private key is required")
	}
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("wg pubkey: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
