// Package postfix manages Postfix mail server configuration.
package postfix

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const defaultMainCf = "/etc/postfix/main.cf"

// Result is the common return type for postfix operations.
type Result struct {
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	Changed    bool   `json:"changed"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Get returns the value of a Postfix configuration parameter.
func Get(key string) (Result, error) {
	start := time.Now()
	if key == "" {
		return Result{Error: "key must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("key must not be empty")
	}

	cmd := exec.Command("postconf", "-h", key)
	out, err := cmd.Output()
	if err != nil {
		// Fallback to parsing main.cf
		val, ferr := getFromFile(defaultMainCf, key)
		if ferr != nil {
			return Result{Key: key, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
		}
		return Result{Key: key, Value: val, DurationMs: time.Since(start).Milliseconds()}, nil
	}
	return Result{Key: key, Value: strings.TrimSpace(string(out)), DurationMs: time.Since(start).Milliseconds()}, nil
}

// Set sets a Postfix configuration parameter (idempotent).
func Set(key, value string) (Result, error) {
	start := time.Now()
	if key == "" {
		return Result{Error: "key must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("key must not be empty")
	}

	// Check current value
	cur, _ := Get(key)
	if cur.Value == value && cur.Error == "" {
		return Result{Key: key, Value: value, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	cmd := exec.Command("postconf", "-e", key+" = "+value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return Result{Key: key, Error: string(out), DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("postconf -e: %s: %w", string(out), err)
	}

	return Result{Key: key, Value: value, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Reload reloads Postfix configuration.
func Reload() (Result, error) {
	start := time.Now()
	cmd := exec.Command("postfix", "reload")
	if out, err := cmd.CombinedOutput(); err != nil {
		return Result{Error: string(out), DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("postfix reload: %s: %w", string(out), err)
	}
	return Result{Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

func getFromFile(path, key string) (string, error) {
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
	return "", fmt.Errorf("parameter %q not found", key)
}
