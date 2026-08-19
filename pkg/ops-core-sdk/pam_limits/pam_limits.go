package pam_limits

import (
	"fmt"
	"os"
	"strings"
)

// LimitResult represents the result of managing a PAM limit.
type LimitResult struct {
	Domain  string `json:"domain"`
	Type    string `json:"type"`
	Item    string `json:"item"`
	Value   string `json:"value"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ListResult represents a list of PAM limits.
type ListResult struct {
	Limits []LimitEntry `json:"limits"`
	Count  int          `json:"count"`
	Error  string       `json:"error,omitempty"`
}

// LimitEntry represents a single PAM limit entry.
type LimitEntry struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Item   string `json:"item"`
	Value  string `json:"value"`
}

const limitsFile = "/etc/security/limits.conf"

// Set sets a PAM limit.
// domain: user, group, or @group
// type: soft, hard, or -
// item: nproc, nofile, core, etc.
// value: the limit value
func Set(domain, limitType, item, value string) LimitResult {
	if domain == "" || limitType == "" || item == "" || value == "" {
		return LimitResult{Error: "domain, type, item, and value are required"}
	}

	// Read current limits
	content, err := os.ReadFile(limitsFile)
	if err != nil {
		return LimitResult{
			Error: fmt.Sprintf("failed to read %s: %s", limitsFile, err),
		}
	}

	newEntry := fmt.Sprintf("%s %s %s %s", domain, limitType, item, value)
	lines := strings.Split(string(content), "\n")
	found := false
	changed := false

	for i, line := range lines {
		// Skip comments and empty lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check if this is the same limit
		fields := strings.Fields(trimmed)
		if len(fields) >= 4 && fields[0] == domain && fields[1] == limitType && fields[2] == item {
			found = true
			if fields[3] != value {
				lines[i] = newEntry
				changed = true
			}
			break
		}
	}

	if !found {
		lines = append(lines, newEntry)
		changed = true
	}

	if changed {
		err = os.WriteFile(limitsFile, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return LimitResult{
				Error: fmt.Sprintf("failed to write %s: %s", limitsFile, err),
			}
		}
	}

	return LimitResult{
		Domain:  domain,
		Type:    limitType,
		Item:    item,
		Value:   value,
		Changed: changed,
	}
}

// List lists all PAM limits.
func List() ListResult {
	content, err := os.ReadFile(limitsFile)
	if err != nil {
		return ListResult{
			Error: fmt.Sprintf("failed to read %s: %s", limitsFile, err),
		}
	}

	var limits []LimitEntry
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) >= 4 {
			limits = append(limits, LimitEntry{
				Domain: fields[0],
				Type:   fields[1],
				Item:   fields[2],
				Value:  fields[3],
			})
		}
	}

	return ListResult{
		Limits: limits,
		Count:  len(limits),
	}
}
