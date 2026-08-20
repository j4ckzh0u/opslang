// Package service_facts provides Ansible service_facts module equivalent.
package service_facts

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
)

// ServiceFactsResult contains discovered services.
type ServiceFactsResult struct {
	Services map[string]ServiceInfo `json:"services"`
	Error    string                 `json:"error,omitempty"`
}

// ServiceInfo describes a service.
type ServiceInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // running, stopped, unknown
	Source  string `json:"source"` // systemd, sysvinit
	Enabled bool   `json:"enabled,omitempty"`
}

// Collect gathers service facts.
func Collect() ServiceFactsResult {
	if runtime.GOOS != "linux" {
		return ServiceFactsResult{Error: "only supported on linux"}
	}
	services := map[string]ServiceInfo{}
	if systemd := collectSystemd(); len(systemd) > 0 {
		for k, v := range systemd {
			services[k] = v
		}
	}
	return ServiceFactsResult{Services: services}
}

func collectSystemd() map[string]ServiceInfo {
	out, err := exec.Command("systemctl", "list-units", "--type=service", "--all", "--plain", "--no-legend").Output()
	if err != nil {
		return nil
	}
	services := map[string]ServiceInfo{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		unit := fields[0]
		if !strings.HasSuffix(unit, ".service") {
			continue
		}
		name := strings.TrimSuffix(unit, ".service")
		load := fields[1]
		active := fields[2]
		status := "unknown"
		if active == "active" {
			status = "running"
		} else if active == "inactive" {
			status = "stopped"
		}
		enabled := false
		if load == "loaded" {
			enabled = isEnabled(name)
		}
		services[name] = ServiceInfo{
			Name:    name,
			Status:  status,
			Source:  "systemd",
			Enabled: enabled,
		}
	}
	return services
}

func isEnabled(name string) bool {
	out, err := exec.Command("systemctl", "is-enabled", name).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "enabled"
}

// JSON returns services as JSON string.
func (r ServiceFactsResult) JSON() string {
	b, _ := json.Marshal(r.Services)
	return string(b)
}
