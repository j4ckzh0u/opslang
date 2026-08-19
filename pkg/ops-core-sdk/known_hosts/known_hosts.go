// Package known_hosts provides SSH known_hosts file management.
package known_hosts

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// Entry represents a known_hosts entry.
type Entry struct {
	Host string `json:"host"`
	Key  string `json:"key"`
}

// ListResult is returned by List.
type ListResult struct {
	Entries []Entry `json:"entries"`
}

// CheckResult is returned by Check.
type CheckResult struct {
	Found bool `json:"found"`
}

// AddResult is returned by Add.
type AddResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

func knownHostsPath() string {
	u, err := user.Current()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
	}
	return filepath.Join(u.HomeDir, ".ssh", "known_hosts")
}

// List returns all known_hosts entries.
func List() (ListResult, error) {
	path := knownHostsPath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ListResult{Entries: []Entry{}}, nil
		}
		return ListResult{}, fmt.Errorf("open known_hosts: %w", err)
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
		if len(fields) < 2 {
			continue
		}
		entries = append(entries, Entry{Host: fields[0], Key: fields[1]})
	}

	return ListResult{Entries: entries}, nil
}

// Check returns whether a host is in known_hosts.
func Check(host string) (CheckResult, error) {
	if host == "" {
		return CheckResult{}, fmt.Errorf("host is required")
	}

	entries, err := List()
	if err != nil {
		return CheckResult{}, err
	}

	for _, e := range entries.Entries {
		for _, h := range strings.Split(e.Host, ",") {
			if h == host {
				return CheckResult{Found: true}, nil
			}
		}
	}

	return CheckResult{Found: false}, nil
}

// Add adds a host key to known_hosts using ssh-keyscan.
func Add(host string) (AddResult, error) {
	if host == "" {
		return AddResult{Error: "host is required"}, fmt.Errorf("host is required")
	}

	result, _ := Check(host)
	if result.Found {
		return AddResult{Changed: false}, nil
	}

	out, err := exec.Command("ssh-keyscan", "-H", host).Output()
	if err != nil {
		return AddResult{Error: fmt.Sprintf("ssh-keyscan failed: %v", err)}, fmt.Errorf("ssh-keyscan: %w", err)
	}

	keyData := strings.TrimSpace(string(out))
	if keyData == "" {
		return AddResult{Error: "no key returned by ssh-keyscan"}, fmt.Errorf("ssh-keyscan returned empty key")
	}

	path := knownHostsPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return AddResult{Error: err.Error()}, fmt.Errorf("mkdir .ssh: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return AddResult{Error: err.Error()}, fmt.Errorf("open known_hosts: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(keyData + "\n"); err != nil {
		return AddResult{Error: err.Error()}, fmt.Errorf("write known_hosts: %w", err)
	}

	return AddResult{Changed: true}, nil
}

// Remove removes a host from known_hosts.
func Remove(host string) (RemoveResult, error) {
	if host == "" {
		return RemoveResult{Error: "host is required"}, fmt.Errorf("host is required")
	}

	path := knownHostsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return RemoveResult{Changed: false}, nil
		}
		return RemoveResult{Error: err.Error()}, fmt.Errorf("read known_hosts: %w", err)
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
		if len(fields) < 1 {
			lines = append(lines, line)
			continue
		}
		match := false
		for _, h := range strings.Split(fields[0], ",") {
			if h == host {
				match = true
				found = true
				break
			}
		}
		if !match {
			lines = append(lines, line)
		}
	}

	if !found {
		return RemoveResult{Changed: false}, nil
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		return RemoveResult{Error: err.Error()}, fmt.Errorf("write known_hosts: %w", err)
	}

	return RemoveResult{Changed: true}, nil
}
