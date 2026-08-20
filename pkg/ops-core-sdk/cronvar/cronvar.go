// Package cronvar provides Ansible cronvar module equivalent.
// Manage variables in crontab files.
package cronvar

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CronvarResult is returned by cronvar operations.
type CronvarResult struct {
	Name    string `json:"name"`
	Value   string `json:"value,omitempty"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// Present ensures a cron variable is set.
func Present(name, value, user string, insertAfter, insertBefore string) CronvarResult {
	if name == "" {
		return CronvarResult{Error: "name is required"}
	}
	if value == "" {
		return CronvarResult{Error: "value is required"}
	}

	crontab := getCrontab(user)
	lines := strings.Split(crontab, "\n")
	entry := name + "=" + value
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, name+"=") {
			if trimmed == entry {
				return CronvarResult{Name: name, Value: value, Changed: false}
			}
			lines[i] = entry
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, entry)
	}
	setCrontab(user, strings.Join(lines, "\n"))
	return CronvarResult{Name: name, Value: value, Changed: true}
}

// Absent removes a cron variable.
func Absent(name, user string) CronvarResult {
	if name == "" {
		return CronvarResult{Error: "name is required"}
	}
	crontab := getCrontab(user)
	lines := strings.Split(crontab, "\n")
	var newLines []string
	removed := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, name+"=") {
			removed = true
			continue
		}
		newLines = append(newLines, line)
	}
	if !removed {
		return CronvarResult{Name: name, Changed: false}
	}
	setCrontab(user, strings.Join(newLines, "\n"))
	return CronvarResult{Name: name, Changed: true}
}

// Get retrieves a cron variable's value.
func Get(name, user string) CronvarResult {
	if name == "" {
		return CronvarResult{Error: "name is required"}
	}
	crontab := getCrontab(user)
	sc := bufio.NewScanner(strings.NewReader(crontab))
	for sc.Scan() {
		trimmed := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(trimmed, name+"=") {
			val := strings.TrimPrefix(trimmed, name+"=")
			return CronvarResult{Name: name, Value: val}
		}
	}
	return CronvarResult{Name: name, Error: "variable not found"}
}

func getCrontab(user string) string {
	args := []string{"-l"}
	if user != "" {
		args = append([]string{"-u", user}, args...)
	}
	out, err := exec.Command("crontab", args...).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func setCrontab(user, content string) {
	tmp, err := os.CreateTemp("", "cronvar-*")
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name())
	fmt.Fprint(tmp, content)
	tmp.Close()
	args := []string{tmp.Name()}
	if user != "" {
		args = append([]string{"-u", user}, args...)
	}
	exec.Command("crontab", args...).Run()
}
