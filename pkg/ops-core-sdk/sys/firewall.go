package sys

import (
	"fmt"
	"os/exec"
	"strings"
)

// FirewallResult is returned by FirewallRule, reporting whether a rule was changed.
type FirewallResult struct {
	Changed bool   `json:"changed"`
	Rule    string `json:"rule"`
	Error   string `json:"error,omitempty"`
}

// FirewallRule adds or removes an iptables INPUT rule.
// action must be "add" or "remove".
// protocol is e.g. "tcp", "udp".
// port is the destination port (1-65535).
// source may be empty or a CIDR/IP (e.g. "10.0.0.0/8").
func FirewallRule(action string, protocol string, port int, source string) (FirewallResult, error) {
	if action != "add" && action != "remove" {
		return FirewallResult{}, fmt.Errorf("failed to manage firewall: action must be \"add\" or \"remove\", got %q", action)
	}
	if protocol == "" {
		return FirewallResult{}, fmt.Errorf("failed to manage firewall: protocol must not be empty")
	}
	if port < 1 || port > 65535 {
		return FirewallResult{}, fmt.Errorf("failed to manage firewall: port must be between 1 and 65535, got %d", port)
	}

	ruleDesc := fmt.Sprintf("%s/%d", protocol, port)
	if source != "" {
		ruleDesc = fmt.Sprintf("%s from %s", ruleDesc, source)
	}

	// Build the check args (iptables -C INPUT ...)
	checkArgs := []string{"-C", "INPUT", "-p", protocol, "--dport", fmt.Sprintf("%d", port)}
	if source != "" {
		checkArgs = append(checkArgs, "-s", source)
	}
	checkArgs = append(checkArgs, "-j", "ACCEPT")

	checkCmd := exec.Command("iptables", checkArgs...)
	checkErr := checkCmd.Run()
	ruleExists := checkErr == nil

	if action == "add" {
		if ruleExists {
			return FirewallResult{Changed: false, Rule: ruleDesc}, nil
		}
		// -A INPUT ...
		addArgs := []string{"-A", "INPUT", "-p", protocol, "--dport", fmt.Sprintf("%d", port)}
		if source != "" {
			addArgs = append(addArgs, "-s", source)
		}
		addArgs = append(addArgs, "-j", "ACCEPT")

		cmd := exec.Command("iptables", addArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return FirewallResult{}, fmt.Errorf("failed to add firewall rule %s: %w: %s", ruleDesc, err, strings.TrimSpace(string(output)))
		}
		return FirewallResult{Changed: true, Rule: ruleDesc}, nil
	}

	// action == "remove"
	if !ruleExists {
		return FirewallResult{Changed: false, Rule: ruleDesc}, nil
	}

	delArgs := []string{"-D", "INPUT", "-p", protocol, "--dport", fmt.Sprintf("%d", port)}
	if source != "" {
		delArgs = append(delArgs, "-s", source)
	}
	delArgs = append(delArgs, "-j", "ACCEPT")

	cmd := exec.Command("iptables", delArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return FirewallResult{}, fmt.Errorf("failed to remove firewall rule %s: %w: %s", ruleDesc, err, strings.TrimSpace(string(output)))
	}
	return FirewallResult{Changed: true, Rule: ruleDesc}, nil
}
