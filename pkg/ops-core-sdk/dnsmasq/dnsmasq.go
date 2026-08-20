// Package dnsmasq manages dnsmasq DNS/DHCP server configuration.
package dnsmasq

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const defaultConfigPath = "/etc/dnsmasq.conf"

// Result is the common return type for dnsmasq operations.
type Result struct {
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	Changed    bool   `json:"changed"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Get returns the value of a dnsmasq.conf directive.
func Get(key string) (Result, error) {
	start := time.Now()
	if key == "" {
		return Result{Error: "key must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("key must not be empty")
	}
	val, err := getDirective(defaultConfigPath, key)
	if err != nil {
		return Result{Key: key, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Key: key, Value: val, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Set sets a directive in dnsmasq.conf (idempotent).
func Set(key, value string) (Result, error) {
	start := time.Now()
	if key == "" {
		return Result{Error: "key must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("key must not be empty")
	}

	cur, err := getDirective(defaultConfigPath, key)
	if err == nil && cur == value {
		return Result{Key: key, Value: value, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	if err := setDirective(defaultConfigPath, key, value); err != nil {
		return Result{Key: key, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Key: key, Value: value, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Absent removes a directive from dnsmasq.conf (idempotent).
func Absent(key string) (Result, error) {
	start := time.Now()
	if key == "" {
		return Result{Error: "key must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("key must not be empty")
	}

	cur, err := getDirective(defaultConfigPath, key)
	if err != nil || cur == "" {
		return Result{Key: key, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	if err := removeDirective(defaultConfigPath, key); err != nil {
		return Result{Key: key, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Key: key, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Restart restarts the dnsmasq service.
func Restart() (Result, error) {
	start := time.Now()
	cmd := exec.Command("systemctl", "restart", "dnsmasq")
	if out, err := cmd.CombinedOutput(); err != nil {
		return Result{Error: string(out), DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("restart dnsmasq: %s: %w", string(out), err)
	}
	return Result{Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

func getDirective(path, key string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			return strings.TrimSpace(parts[1]), nil
		}
	}
	return "", fmt.Errorf("directive %q not found", key)
}

func setDirective(path, key, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) >= 1 && strings.TrimSpace(parts[0]) == key {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, key+"="+value)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func removeDirective(path, key string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) >= 1 && strings.TrimSpace(parts[0]) == key {
			continue
		}
		out = append(out, line)
	}

	return os.WriteFile(path, []byte(strings.Join(out, "\n")), 0644)
}
