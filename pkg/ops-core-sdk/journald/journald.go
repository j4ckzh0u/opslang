// Package journald manages systemd journal configuration.
package journald

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const defaultConfigPath = "/etc/systemd/journald.conf"

// Result is the common return type for journald operations.
type Result struct {
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	Changed    bool   `json:"changed"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Get returns the value of a journald.conf directive.
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

// Set sets a directive in journald.conf (idempotent).
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

func getDirective(path, key string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inJournal := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inJournal = strings.EqualFold(line, "[Journal]")
			continue
		}
		if !inJournal {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), key) {
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
	inJournal := false
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			inJournal = strings.EqualFold(trimmed, "[Journal]")
			continue
		}
		if !inJournal {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), key) {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}

	if !found {
		// Append to [Journal] section or create it
		journalIdx := -1
		for i, line := range lines {
			if strings.TrimSpace(line) == "[Journal]" {
				journalIdx = i
				break
			}
		}
		if journalIdx >= 0 {
			// Insert after [Journal] header
			newLine := key + "=" + value
			lines = append(lines[:journalIdx+1], append([]string{newLine}, lines[journalIdx+1:]...)...)
		} else {
			lines = append(lines, "[Journal]", key+"="+value)
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
