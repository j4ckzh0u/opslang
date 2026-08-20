// Package sshd_config manages SSH daemon configuration.
package sshd_config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

const defaultConfigPath = "/etc/ssh/sshd_config"

// Result is the common return type for sshd_config operations.
type Result struct {
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	Changed    bool   `json:"changed"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Get returns the value of an sshd_config directive.
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

// Set sets a directive in sshd_config (idempotent).
func Set(key, value string) (Result, error) {
	start := time.Now()
	if key == "" {
		return Result{Error: "key must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("key must not be empty")
	}
	if value == "" {
		return Result{Error: "value must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("value must not be empty")
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

// Absent removes a directive from sshd_config (idempotent).
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

func getDirective(path, key string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	prefix := strings.ToLower(key)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.ToLower(fields[0]) == prefix {
			return strings.Join(fields[1:], " "), nil
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
	prefix := strings.ToLower(key)
	re := regexp.MustCompile(`(?i)^\s*` + regexp.QuoteMeta(key) + `\s+`)
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 1 && strings.ToLower(fields[0]) == prefix {
			lines[i] = key + " " + value
			found = true
			break
		}
	}

	if !found {
		_ = re // re used for potential future use
		lines = append(lines, key+" "+value)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func removeDirective(path, key string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	prefix := strings.ToLower(key)
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 1 && strings.ToLower(fields[0]) == prefix {
			continue
		}
		out = append(out, line)
	}

	return os.WriteFile(path, []byte(strings.Join(out, "\n")), 0644)
}
