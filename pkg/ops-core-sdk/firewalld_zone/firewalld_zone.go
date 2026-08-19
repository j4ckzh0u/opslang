// Package firewalld_zone provides firewalld zone and service/port management.
// Manages zones, services, ports, rich rules, and default zone configuration.
package firewalld_zone

import (
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a firewalld action.
type ActionResult struct {
	Zone    string `json:"zone,omitempty"`
	Action  string `json:"action"`
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
}

// ZoneInfo represents information about a firewalld zone.
type ZoneInfo struct {
	Name      string   `json:"name"`
	Default   bool     `json:"default"`
	Services  []string `json:"services"`
	Ports     []string `json:"ports"`
	Protocols []string `json:"protocols"`
	Sources   []string `json:"sources"`
	Interfaces []string `json:"interfaces"`
	RichRules []string `json:"rich_rules"`
	Target    string   `json:"target"`
}

// ListResult represents the result of listing zones.
type ListResult struct {
	Zones []string `json:"zones"`
	Count int      `json:"count"`
}

// runFirewallCmd executes a firewall-cmd command.
func runFirewallCmd(args ...string) (string, error) {
	cmd := exec.Command("firewall-cmd", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// GetDefaultZone returns the current default zone.
func GetDefaultZone() (*ActionResult, error) {
	result := &ActionResult{Action: "get_default_zone"}
	out, err := runFirewallCmd("--get-default-zone")
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.GetDefaultZone: %w (%s)", err, out)
	}
	result.Zone = out
	result.Message = out
	return result, nil
}

// SetDefaultZone sets the default zone.
func SetDefaultZone(zone string) (*ActionResult, error) {
	if zone == "" {
		return nil, fmt.Errorf("firewalld_zone.SetDefaultZone: zone is required")
	}
	result := &ActionResult{Zone: zone, Action: "set_default_zone"}

	current, _ := runFirewallCmd("--get-default-zone")
	if current == zone {
		result.Changed = false
		result.Message = "already default"
		return result, nil
	}

	out, err := runFirewallCmd("--set-default-zone="+zone)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.SetDefaultZone: %w (%s)", err, out)
	}
	result.Changed = true
	result.Message = "default zone set to " + zone
	return result, nil
}

// AddZone adds a new permanent zone.
func AddZone(zone string) (*ActionResult, error) {
	if zone == "" {
		return nil, fmt.Errorf("firewalld_zone.AddZone: zone is required")
	}
	result := &ActionResult{Zone: zone, Action: "add_zone"}

	// Check if exists
	out, _ := runFirewallCmd("--get-zones")
	zones := strings.Fields(out)
	for _, z := range zones {
		if z == zone {
			result.Changed = false
			result.Message = "zone already exists"
			return result, nil
		}
	}

	out, err := runFirewallCmd("--permanent", "--new-zone="+zone)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.AddZone: %w (%s)", err, out)
	}
	// Reload to apply
	runFirewallCmd("--reload")
	result.Changed = true
	result.Message = "zone created"
	return result, nil
}

// RemoveZone removes a permanent zone.
func RemoveZone(zone string) (*ActionResult, error) {
	if zone == "" {
		return nil, fmt.Errorf("firewalld_zone.RemoveZone: zone is required")
	}
	result := &ActionResult{Zone: zone, Action: "remove_zone"}

	out, err := runFirewallCmd("--permanent", "--delete-zone="+zone)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.RemoveZone: %w (%s)", err, out)
	}
	runFirewallCmd("--reload")
	result.Changed = true
	result.Message = "zone removed"
	return result, nil
}

// AddService adds a service to a zone permanently.
func AddService(zone string, service string) (*ActionResult, error) {
	if zone == "" || service == "" {
		return nil, fmt.Errorf("firewalld_zone.AddService: zone and service are required")
	}
	result := &ActionResult{Zone: zone, Action: "add_service"}

	// Check if already added
	out, _ := runFirewallCmd("--permanent", "--zone="+zone, "--list-services")
	services := strings.Fields(out)
	for _, s := range services {
		if s == service {
			result.Changed = false
			result.Message = "service already in zone"
			return result, nil
		}
	}

	out, err := runFirewallCmd("--permanent", "--zone="+zone, "--add-service="+service)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.AddService: %w (%s)", err, out)
	}
	runFirewallCmd("--reload")
	result.Changed = true
	result.Message = fmt.Sprintf("service %s added to zone %s", service, zone)
	return result, nil
}

// RemoveService removes a service from a zone.
func RemoveService(zone string, service string) (*ActionResult, error) {
	if zone == "" || service == "" {
		return nil, fmt.Errorf("firewalld_zone.RemoveService: zone and service are required")
	}
	result := &ActionResult{Zone: zone, Action: "remove_service"}

	out, err := runFirewallCmd("--permanent", "--zone="+zone, "--remove-service="+service)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.RemoveService: %w (%s)", err, out)
	}
	runFirewallCmd("--reload")
	result.Changed = true
	result.Message = fmt.Sprintf("service %s removed from zone %s", service, zone)
	return result, nil
}

// AddPort adds a port/protocol to a zone permanently.
// portProtocol format: "port/protocol" (e.g., "8080/tcp", "53/udp").
func AddPort(zone string, portProtocol string) (*ActionResult, error) {
	if zone == "" || portProtocol == "" {
		return nil, fmt.Errorf("firewalld_zone.AddPort: zone and port/protocol are required")
	}
	result := &ActionResult{Zone: zone, Action: "add_port"}

	// Check if already added
	out, _ := runFirewallCmd("--permanent", "--zone="+zone, "--list-ports")
	ports := strings.Fields(out)
	for _, p := range ports {
		if p == portProtocol {
			result.Changed = false
			result.Message = "port already in zone"
			return result, nil
		}
	}

	out, err := runFirewallCmd("--permanent", "--zone="+zone, "--add-port="+portProtocol)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.AddPort: %w (%s)", err, out)
	}
	runFirewallCmd("--reload")
	result.Changed = true
	result.Message = fmt.Sprintf("port %s added to zone %s", portProtocol, zone)
	return result, nil
}

// RemovePort removes a port/protocol from a zone.
func RemovePort(zone string, portProtocol string) (*ActionResult, error) {
	if zone == "" || portProtocol == "" {
		return nil, fmt.Errorf("firewalld_zone.RemovePort: zone and port/protocol are required")
	}
	result := &ActionResult{Zone: zone, Action: "remove_port"}

	out, err := runFirewallCmd("--permanent", "--zone="+zone, "--remove-port="+portProtocol)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.RemovePort: %w (%s)", err, out)
	}
	runFirewallCmd("--reload")
	result.Changed = true
	result.Message = fmt.Sprintf("port %s removed from zone %s", portProtocol, zone)
	return result, nil
}

// AddRichRule adds a rich language rule to a zone.
func AddRichRule(zone string, rule string) (*ActionResult, error) {
	if zone == "" || rule == "" {
		return nil, fmt.Errorf("firewalld_zone.AddRichRule: zone and rule are required")
	}
	result := &ActionResult{Zone: zone, Action: "add_rich_rule"}

	out, err := runFirewallCmd("--permanent", "--zone="+zone, "--add-rich-rule="+rule)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.AddRichRule: %w (%s)", err, out)
	}
	runFirewallCmd("--reload")
	result.Changed = true
	result.Message = "rich rule added"
	return result, nil
}

// RemoveRichRule removes a rich language rule from a zone.
func RemoveRichRule(zone string, rule string) (*ActionResult, error) {
	if zone == "" || rule == "" {
		return nil, fmt.Errorf("firewalld_zone.RemoveRichRule: zone and rule are required")
	}
	result := &ActionResult{Zone: zone, Action: "remove_rich_rule"}

	out, err := runFirewallCmd("--permanent", "--zone="+zone, "--remove-rich-rule="+rule)
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.RemoveRichRule: %w (%s)", err, out)
	}
	runFirewallCmd("--reload")
	result.Changed = true
	result.Message = "rich rule removed"
	return result, nil
}

// Info returns detailed information about a zone.
func Info(zone string) (*ZoneInfo, error) {
	if zone == "" {
		return nil, fmt.Errorf("firewalld_zone.Info: zone is required")
	}
	info := &ZoneInfo{Name: zone}

	// Check if this is the default zone
	defaultZone, _ := runFirewallCmd("--get-default-zone")
	info.Default = (defaultZone == zone)

	// Get services
	servicesOut, _ := runFirewallCmd("--permanent", "--zone="+zone, "--list-services")
	if servicesOut != "" {
		info.Services = strings.Fields(servicesOut)
	}

	// Get ports
	portsOut, _ := runFirewallCmd("--permanent", "--zone="+zone, "--list-ports")
	if portsOut != "" {
		info.Ports = strings.Fields(portsOut)
	}

	// Get protocols
	protoOut, _ := runFirewallCmd("--permanent", "--zone="+zone, "--list-protocols")
	if protoOut != "" {
		info.Protocols = strings.Fields(protoOut)
	}

	// Get sources
	srcOut, _ := runFirewallCmd("--permanent", "--zone="+zone, "--list-sources")
	if srcOut != "" {
		info.Sources = strings.Fields(srcOut)
	}

	// Get interfaces
	ifaceOut, _ := runFirewallCmd("--permanent", "--zone="+zone, "--list-interfaces")
	if ifaceOut != "" {
		info.Interfaces = strings.Fields(ifaceOut)
	}

	// Get target
	target, _ := runFirewallCmd("--permanent", "--zone="+zone, "--get-target")
	info.Target = target

	return info, nil
}

// ListZones returns all available zones.
func ListZones() (*ListResult, error) {
	result := &ListResult{}
	out, err := runFirewallCmd("--get-zones")
	if err != nil {
		return nil, fmt.Errorf("firewalld_zone.ListZones: %w (%s)", err, out)
	}
	result.Zones = strings.Fields(out)
	result.Count = len(result.Zones)
	return result, nil
}
