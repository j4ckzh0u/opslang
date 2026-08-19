// Package cron provides crontab management operations (list, add, remove entries).
// All functions use the crontab command directly via os/exec (no shell invocation).
package cron

import (
	"fmt"
	"os/exec"
	"strings"
)

// crontabBin is the path to the crontab binary.
// It is a variable so tests can override it.
var crontabBin = "crontab"

// CronEntry represents a single crontab entry.
type CronEntry struct {
	Minute     string `json:"minute"`
	Hour       string `json:"hour"`
	DayOfMonth string `json:"day_of_month"`
	Month      string `json:"month"`
	DayOfWeek  string `json:"day_of_week"`
	Command    string `json:"command"`
	User       string `json:"user,omitempty"`
}

// ListResult holds the entries returned by List.
type ListResult struct {
	Entries []CronEntry `json:"entries"`
}

// AddResult holds the result of an Add operation.
type AddResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult holds the result of a Remove operation.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// getRawCrontab returns the raw output of crontab -l -u <user>.
// If the user has no crontab, it returns an empty string and nil error.
func getRawCrontab(user string) (string, error) {
	cmd := exec.Command(crontabBin, "-l", "-u", user)
	out, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		// "no crontab for <user>" means empty crontab, not an error.
		if strings.Contains(output, "no crontab for") {
			return "", nil
		}
		return "", fmt.Errorf("crontab -l -u %q failed: %w: %s", user, err, output)
	}
	return string(out), nil
}

// parseLine attempts to parse a single crontab line into a CronEntry.
// It returns the entry and true on success, or false if the line is a
// comment, blank, or cannot be parsed.
func parseLine(line string) (CronEntry, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return CronEntry{}, false
	}

	// A valid cron line has at least 6 whitespace-separated tokens:
	// 5 time fields + at least one token for the command.
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return CronEntry{}, false
	}

	return CronEntry{
		Minute:     fields[0],
		Hour:       fields[1],
		DayOfMonth: fields[2],
		Month:      fields[3],
		DayOfWeek:  fields[4],
		Command:    strings.Join(fields[5:], " "),
	}, true
}

// List returns all crontab entries for the given user.
func List(user string) (ListResult, error) {
	raw, err := getRawCrontab(user)
	if err != nil {
		return ListResult{}, err
	}

	result := ListResult{}
	for _, line := range strings.Split(raw, "\n") {
		entry, ok := parseLine(line)
		if !ok {
			continue
		}
		entry.User = user
		result.Entries = append(result.Entries, entry)
	}
	return result, nil
}

// entryToLine converts a CronEntry back into a crontab line.
func entryToLine(e CronEntry) string {
	return fmt.Sprintf("%s %s %s %s %s %s",
		e.Minute, e.Hour, e.DayOfMonth, e.Month, e.DayOfWeek, e.Command)
}

// replaceCrontab replaces the user's crontab with the provided content.
func replaceCrontab(user, content string) error {
	cmd := exec.Command(crontabBin, "-u", user, "-")
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("crontab -u %q - failed: %w: %s", user, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Add appends a new cron entry to the user's crontab.
func Add(user string, entry CronEntry) (AddResult, error) {
	result := AddResult{}

	raw, err := getRawCrontab(user)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	newLine := entryToLine(entry)

	// Build the new crontab content.
	var builder strings.Builder
	if raw != "" {
		builder.WriteString(strings.TrimRight(raw, "\n"))
		builder.WriteString("\n")
	}
	builder.WriteString(newLine)
	builder.WriteString("\n")

	if err := replaceCrontab(user, builder.String()); err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Changed = true
	return result, nil
}

// Remove filters out all lines containing lineMatch from the user's crontab.
// Returns Changed=true if any lines were removed.
func Remove(user string, lineMatch string) (RemoveResult, error) {
	result := RemoveResult{}

	raw, err := getRawCrontab(user)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	var kept []string
	changed := false
	for _, line := range strings.Split(raw, "\n") {
		if strings.Contains(line, lineMatch) {
			changed = true
			continue
		}
		kept = append(kept, line)
	}

	if !changed {
		return result, nil
	}

	// Rebuild the crontab content from the kept lines.
	content := strings.Join(kept, "\n")
	// Ensure the crontab ends with a newline if there is any content.
	if strings.TrimSpace(content) != "" {
		content = strings.TrimRight(content, "\n") + "\n"
	}

	if err := replaceCrontab(user, content); err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Changed = true
	return result, nil
}
