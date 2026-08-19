// Package firewalld_rich_rule manages firewalld rich rules.
// Equivalent to ansible.posix.firewalld_rich_rule module.
package firewalld_rich_rule

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Rule    string `json:"rule,omitempty"`
	Zone    string `json:"zone,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Status string   `json:"status"`
	Zone   string   `json:"zone"`
	Rules  []string `json:"rules"`
	Count  int      `json:"count"`
	Error  string   `json:"error,omitempty"`
}

// Add adds a rich rule to a zone.
func Add(zone, rule string) Result {
	if rule == "" {
		return Result{Status: "failed", Error: "rule is required"}
	}
	if zone == "" {
		zone = "public"
	}

	args := []string{"--permanent", "--zone=" + zone, "--add-rich-rule=" + rule}
	cmd := exec.Command("firewall-cmd", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Zone: zone, Rule: rule,
			Error: fmt.Sprintf("firewall-cmd: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	// Reload to apply
	exec.Command("firewall-cmd", "--reload").Run()

	return Result{Status: "success", Changed: true, Zone: zone, Rule: rule}
}

// Remove removes a rich rule from a zone.
func Remove(zone, rule string) Result {
	if rule == "" {
		return Result{Status: "failed", Error: "rule is required"}
	}
	if zone == "" {
		zone = "public"
	}

	args := []string{"--permanent", "--zone=" + zone, "--remove-rich-rule=" + rule}
	cmd := exec.Command("firewall-cmd", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Zone: zone, Rule: rule,
			Error: fmt.Sprintf("firewall-cmd: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	exec.Command("firewall-cmd", "--reload").Run()

	return Result{Status: "success", Changed: true, Zone: zone, Rule: rule}
}

// List lists all rich rules in a zone.
func List(zone string) ListResult {
	if zone == "" {
		zone = "public"
	}

	cmd := exec.Command("firewall-cmd", "--permanent", "--zone="+zone, "--list-rich-rules")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Zone: zone,
			Error: fmt.Sprintf("firewall-cmd: %v", err)}
	}

	var rules []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			rules = append(rules, line)
		}
	}

	return ListResult{
		Status: "success",
		Zone:   zone,
		Rules:  rules,
		Count:  len(rules),
	}
}

// Exists checks if a rich rule exists in a zone.
func Exists(zone, rule string) Result {
	if rule == "" {
		return Result{Status: "failed", Error: "rule is required"}
	}
	if zone == "" {
		zone = "public"
	}

	cmd := exec.Command("firewall-cmd", "--permanent", "--zone="+zone, "--query-rich-rule="+rule)
	err := cmd.Run()
	if err != nil {
		return Result{Status: "success", Changed: false, Zone: zone, Rule: rule}
	}
	return Result{Status: "success", Changed: true, Zone: zone, Rule: rule}
}
