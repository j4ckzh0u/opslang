// Package ssh_config manages SSH client configuration.
package ssh_config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Result is the common return type for ssh_config operations.
type Result struct {
	Host         string `json:"host"`
	Option       string `json:"option,omitempty"`
	Value        string `json:"value,omitempty"`
	Changed      bool   `json:"changed"`
	Error        string `json:"error,omitempty"`
	DurationMs   int64  `json:"duration_ms"`
}

const (
	defaultUserConfig   = "~/.ssh/config"
	defaultSystemConfig = "/etc/ssh/ssh_config"
)

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// Get returns the value of an SSH config option for a host.
func Get(host, option, scope string) (Result, error) {
	start := time.Now()
	if host == "" {
		return Result{Error: "host must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("host must not be empty")
	}
	if option == "" {
		return Result{Error: "option must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("option must not be empty")
	}

	path := defaultUserConfig
	if scope == "system" {
		path = defaultSystemConfig
	}
	path = expandPath(path)

	val, err := getOption(path, host, option)
	if err != nil {
		return Result{Host: host, Option: option, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Host: host, Option: option, Value: val, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Set sets an SSH config option for a host (idempotent).
func Set(host, option, value, scope string) (Result, error) {
	start := time.Now()
	if host == "" {
		return Result{Error: "host must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("host must not be empty")
	}
	if option == "" {
		return Result{Error: "option must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("option must not be empty")
	}
	if value == "" {
		return Result{Error: "value must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("value must not be empty")
	}

	path := defaultUserConfig
	if scope == "system" {
		path = defaultSystemConfig
	}
	path = expandPath(path)

	cur, _ := getOption(path, host, option)
	if cur == value {
		return Result{Host: host, Option: option, Value: value, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	if err := setOption(path, host, option, value); err != nil {
		return Result{Host: host, Option: option, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Host: host, Option: option, Value: value, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Absent removes an SSH config option for a host (idempotent).
func Absent(host, option, scope string) (Result, error) {
	start := time.Now()
	if host == "" {
		return Result{Error: "host must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("host must not be empty")
	}
	if option == "" {
		return Result{Error: "option must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("option must not be empty")
	}

	path := defaultUserConfig
	if scope == "system" {
		path = defaultSystemConfig
	}
	path = expandPath(path)

	cur, _ := getOption(path, host, option)
	if cur == "" {
		return Result{Host: host, Option: option, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	if err := removeOption(path, host, option); err != nil {
		return Result{Host: host, Option: option, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Host: host, Option: option, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

func getOption(path, host, option string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found: %s", path)
		}
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	inHost := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		if key == "host" {
			inHost = strings.EqualFold(val, host)
			continue
		}

		if inHost && key == strings.ToLower(option) {
			return val, nil
		}
	}
	return "", fmt.Errorf("option %q not found for host %q", option, host)
}

func setOption(path, host, option, value string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	if len(data) == 0 {
		// Create new file
		content := fmt.Sprintf("Host %s\n    %s %s\n", host, option, value)
		return os.WriteFile(path, []byte(content), 0600)
	}

	lines := strings.Split(string(data), "\n")
	inHost := false
	hostIdx := -1
	optionIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		if key == "host" && strings.EqualFold(val, host) {
			inHost = true
			hostIdx = i
			continue
		}

		if inHost {
			if key == "host" {
				// Reached next host block
				break
			}
			if key == strings.ToLower(option) {
				optionIdx = i
				break
			}
		}
	}

	if hostIdx < 0 {
		// Add new host block
		lines = append(lines, fmt.Sprintf("Host %s", host), fmt.Sprintf("    %s %s", option, value))
	} else if optionIdx >= 0 {
		// Update existing option
		lines[optionIdx] = fmt.Sprintf("    %s %s", option, value)
	} else {
		// Add option to existing host block
		insertIdx := hostIdx + 1
		for insertIdx < len(lines) {
			trimmed := strings.TrimSpace(lines[insertIdx])
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				insertIdx++
				continue
			}
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) >= 2 && strings.ToLower(strings.TrimSpace(parts[0])) == "host" {
				break
			}
			insertIdx++
		}
		lines = append(lines[:insertIdx], append([]string{fmt.Sprintf("    %s %s", option, value)}, lines[insertIdx:]...)...)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}

func removeOption(path, host, option string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	inHost := false
	optionIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		if key == "host" {
			inHost = strings.EqualFold(val, host)
			continue
		}

		if inHost && key == strings.ToLower(option) {
			optionIdx = i
			break
		}
	}

	if optionIdx < 0 {
		return nil
	}

	lines = append(lines[:optionIdx], lines[optionIdx+1:]...)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}
