// Package ufw provides Ubuntu firewall (UFW) management operations.
package ufw

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// StatusResult represents UFW status.
type StatusResult struct {
	Status   string `json:"status"`    // active/inactive
	Logging  string `json:"logging"`   // off/low/medium/high
	Profiles string `json:"profiles"`  // default profile
}

// Rule represents a UFW rule.
type Rule struct {
	Action   string `json:"action"`   // ALLOW/DENY/REJECT/LIMIT
	To       string `json:"to"`       // destination
	From     string `json:"from"`     // source
}

// ListResult represents the result of listing rules.
type ListResult struct {
	Rules []Rule `json:"rules"`
}

// ActionResult represents the result of a UFW action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// Status returns UFW status.
func Status() (*StatusResult, error) {
	out, err := exec.Command("ufw", "status", "verbose").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ufw status failed: %w", err)
	}

	result := &StatusResult{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Status:") {
			result.Status = strings.TrimSpace(strings.TrimPrefix(line, "Status:"))
		} else if strings.HasPrefix(line, "Logging:") {
			result.Logging = strings.TrimSpace(strings.TrimPrefix(line, "Logging:"))
		}
	}

	return result, nil
}

// List returns all UFW rules.
func List() (*ListResult, error) {
	out, err := exec.Command("ufw", "status", "numbered").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ufw status failed: %w", err)
	}

	result := &ListResult{Rules: make([]Rule, 0)}
	lines := strings.Split(string(out), "\n")

	// Parse numbered rules
	re := regexp.MustCompile(`\[\s*(\d+)\]\s+(\S+)\s+(.+)`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 4 {
			rule := Rule{
				Action: matches[2],
			}
			// Parse "To" and "From" from the rest
			parts := strings.Split(matches[3], " ")
			if len(parts) >= 1 {
				rule.To = parts[0]
			}
			if len(parts) >= 3 && parts[1] == "from" {
				rule.From = parts[2]
			}
			result.Rules = append(result.Rules, rule)
		}
	}

	return result, nil
}

// Enable enables UFW.
func Enable() (*ActionResult, error) {
	cmd := exec.Command("ufw", "--force", "enable")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ufw enable failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: "UFW enabled",
	}, nil
}

// Disable disables UFW.
func Disable() (*ActionResult, error) {
	cmd := exec.Command("ufw", "--force", "disable")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ufw disable failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: "UFW disabled",
	}, nil
}

// Allow adds an allow rule.
func Allow(port string, proto string) (*ActionResult, error) {
	if proto == "" {
		proto = "tcp"
	}

	cmd := exec.Command("ufw", "allow", fmt.Sprintf("%s/%s", port, proto))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ufw allow failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Allowed %s/%s", port, proto),
	}, nil
}

// Deny adds a deny rule.
func Deny(port string, proto string) (*ActionResult, error) {
	if proto == "" {
		proto = "tcp"
	}

	cmd := exec.Command("ufw", "deny", fmt.Sprintf("%s/%s", port, proto))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ufw deny failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Denied %s/%s", port, proto),
	}, nil
}

// Delete deletes a rule by number.
func Delete(number int) (*ActionResult, error) {
	cmd := exec.Command("ufw", "--force", "delete", strconv.Itoa(number))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ufw delete failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Deleted rule #%d", number),
	}, nil
}

// Reset resets UFW to default settings.
func Reset() (*ActionResult, error) {
	cmd := exec.Command("ufw", "--force", "reset")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ufw reset failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: "UFW reset to defaults",
	}, nil
}

// Reload reloads UFW rules.
func Reload() (*ActionResult, error) {
	cmd := exec.Command("ufw", "reload")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ufw reload failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: "UFW reloaded",
	}, nil
}
