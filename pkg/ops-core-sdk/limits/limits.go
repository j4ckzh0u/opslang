// Package limits provides /etc/security/limits.conf management.
package limits

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Entry represents a single limits.conf entry.
type Entry struct {
	Domain string `json:"domain"` // username, @group, or *
	Type   string `json:"type"`   // soft, hard, -
	Item   string `json:"item"`   // core, data, nofile, nproc, etc.
	Value  string `json:"value"`
}

// ListResult is returned by List.
type ListResult struct {
	Entries []Entry `json:"entries"`
}

// GetResult is returned by Get.
type GetResult struct {
	Entries []Entry `json:"entries"`
}

// SetResult is returned by Set.
type SetResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

const limitsConf = "/etc/security/limits.conf"

// List returns all entries from limits.conf.
func List() (ListResult, error) {
	f, err := os.Open(limitsConf)
	if err != nil {
		return ListResult{}, fmt.Errorf("open limits.conf: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		entries = append(entries, Entry{
			Domain: fields[0],
			Type:   fields[1],
			Item:   fields[2],
			Value:  fields[3],
		})
	}

	return ListResult{Entries: entries}, nil
}

// Get returns entries matching a specific domain.
func Get(domain string) (GetResult, error) {
	if domain == "" {
		return GetResult{}, fmt.Errorf("domain is required")
	}

	all, err := List()
	if err != nil {
		return GetResult{}, err
	}

	var entries []Entry
	for _, e := range all.Entries {
		if e.Domain == domain {
			entries = append(entries, e)
		}
	}

	return GetResult{Entries: entries}, nil
}

// Set adds or updates a limits entry.
func Set(domain, typ, item, value string) (SetResult, error) {
	if domain == "" || typ == "" || item == "" || value == "" {
		return SetResult{Error: "all parameters required"}, fmt.Errorf("all parameters required")
	}

	data, err := os.ReadFile(limitsConf)
	if err != nil {
		return SetResult{Error: err.Error()}, fmt.Errorf("read limits.conf: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 4 && fields[0] == domain && fields[1] == typ && fields[2] == item {
			lines[i] = fmt.Sprintf("%s\t%s\t%s\t%s", domain, typ, item, value)
			found = true
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", domain, typ, item, value))
	}

	if err := os.WriteFile(limitsConf, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return SetResult{Error: err.Error()}, fmt.Errorf("write limits.conf: %w", err)
	}

	return SetResult{Changed: true}, nil
}

// Remove removes all entries matching a domain.
func Remove(domain string) (RemoveResult, error) {
	if domain == "" {
		return RemoveResult{Error: "domain is required"}, fmt.Errorf("domain is required")
	}

	data, err := os.ReadFile(limitsConf)
	if err != nil {
		return RemoveResult{Error: err.Error()}, fmt.Errorf("read limits.conf: %w", err)
	}

	var lines []string
	found := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 1 && fields[0] == domain {
			found = true
			continue
		}
		lines = append(lines, line)
	}

	if !found {
		return RemoveResult{Changed: false}, nil
	}

	if err := os.WriteFile(limitsConf, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return RemoveResult{Error: err.Error()}, fmt.Errorf("write limits.conf: %w", err)
	}

	return RemoveResult{Changed: true}, nil
}
