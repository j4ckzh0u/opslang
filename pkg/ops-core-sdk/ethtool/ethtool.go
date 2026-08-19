// Package ethtool provides network interface configuration management.
package ethtool

import (
	"fmt"
	"os/exec"
	"strings"
)

// InterfaceInfo represents network interface settings.
type InterfaceInfo struct {
	Interface string `json:"interface"`
	Speed     string `json:"speed,omitempty"`
	Duplex    string `json:"duplex,omitempty"`
	Link      string `json:"link,omitempty"`
	AutoNeg   string `json:"autoneg,omitempty"`
	Driver    string `json:"driver,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Result is returned by interface operations.
type Result struct {
	Interface string `json:"interface,omitempty"`
	Success   bool   `json:"success"`
	Changed   bool   `json:"changed"`
	Error     string `json:"error,omitempty"`
}

func ethtool(args ...string) (string, error) {
	cmd := exec.Command("ethtool", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Show returns interface information.
func Show(iface string) InterfaceInfo {
	if iface == "" {
		return InterfaceInfo{Error: "interface is required"}
	}
	out, err := ethtool(iface)
	if err != nil {
		return InterfaceInfo{Interface: iface, Error: fmt.Sprintf("ethtool failed: %s: %s", err, out)}
	}
	info := InterfaceInfo{Interface: iface}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Speed:") {
			info.Speed = strings.TrimPrefix(line, "Speed: ")
		} else if strings.HasPrefix(line, "Duplex:") {
			info.Duplex = strings.TrimPrefix(line, "Duplex: ")
		} else if strings.HasPrefix(line, "Link detected:") {
			info.Link = strings.TrimPrefix(line, "Link detected: ")
		} else if strings.HasPrefix(line, "Auto-negotiation:") {
			info.AutoNeg = strings.TrimPrefix(line, "Auto-negotiation: ")
		} else if strings.HasPrefix(line, "driver:") {
			info.Driver = strings.TrimPrefix(line, "driver: ")
		}
	}
	return info
}

// SetSpeed sets interface speed.
func SetSpeed(iface, speed string) Result {
	if iface == "" || speed == "" {
		return Result{Error: "interface and speed are required"}
	}
	out, err := ethtool("-s", iface, "speed", speed)
	if err != nil {
		return Result{Interface: iface, Error: fmt.Sprintf("set speed failed: %s: %s", err, out)}
	}
	return Result{Interface: iface, Success: true, Changed: true}
}

// SetDuplex sets interface duplex mode.
func SetDuplex(iface, duplex string) Result {
	if iface == "" || duplex == "" {
		return Result{Error: "interface and duplex are required"}
	}
	out, err := ethtool("-s", iface, "duplex", duplex)
	if err != nil {
		return Result{Interface: iface, Error: fmt.Sprintf("set duplex failed: %s: %s", err, out)}
	}
	return Result{Interface: iface, Success: true, Changed: true}
}

// SetAutoneg sets auto-negotiation.
func SetAutoneg(iface, autoneg string) Result {
	if iface == "" || autoneg == "" {
		return Result{Error: "interface and autoneg are required"}
	}
	out, err := ethtool("-s", iface, "autoneg", autoneg)
	if err != nil {
		return Result{Interface: iface, Error: fmt.Sprintf("set autoneg failed: %s: %s", err, out)}
	}
	return Result{Interface: iface, Success: true, Changed: true}
}

// SetPause sets pause parameters.
func SetPause(iface, rx, tx string) Result {
	if iface == "" {
		return Result{Error: "interface is required"}
	}
	args := []string{"-A", iface}
	if rx != "" {
		args = append(args, "rx", rx)
	}
	if tx != "" {
		args = append(args, "tx", tx)
	}
	out, err := ethtool(args...)
	if err != nil {
		return Result{Interface: iface, Error: fmt.Sprintf("set pause failed: %s: %s", err, out)}
	}
	return Result{Interface: iface, Success: true, Changed: true}
}

// SetOffload sets offload parameters.
func SetOffload(iface, feature, value string) Result {
	if iface == "" || feature == "" || value == "" {
		return Result{Error: "interface, feature, and value are required"}
	}
	out, err := ethtool("-K", iface, feature, value)
	if err != nil {
		return Result{Interface: iface, Error: fmt.Sprintf("set offload failed: %s: %s", err, out)}
	}
	return Result{Interface: iface, Success: true, Changed: true}
}
