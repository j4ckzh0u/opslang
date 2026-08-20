// Package apache2_module manages Apache HTTP Server modules.
package apache2_module

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Result is the common return type for apache2_module operations.
type Result struct {
	Module     string `json:"module"`
	Enabled    bool   `json:"enabled,omitempty"`
	Changed    bool   `json:"changed"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Check returns whether an Apache module is enabled.
func Check(module string) (Result, error) {
	start := time.Now()
	if module == "" {
		return Result{Error: "module must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("module must not be empty")
	}

	cmd := exec.Command("apache2ctl", "-M")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Try httpd as fallback
		cmd = exec.Command("httpd", "-M")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return Result{Module: module, Error: string(out), DurationMs: time.Since(start).Milliseconds()}, err
		}
	}

	enabled := false
	modName := strings.ToLower(module) + "_module"
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(strings.ToLower(line), modName) {
			enabled = true
			break
		}
	}

	return Result{Module: module, Enabled: enabled, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Enable enables an Apache module (idempotent).
func Enable(module string) (Result, error) {
	start := time.Now()
	if module == "" {
		return Result{Error: "module must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("module must not be empty")
	}

	// Check if already enabled
	check, _ := Check(module)
	if check.Enabled {
		return Result{Module: module, Enabled: true, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	cmd := exec.Command("a2enmod", module)
	if out, err := cmd.CombinedOutput(); err != nil {
		return Result{Module: module, Error: string(out), DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("a2enmod: %s: %w", string(out), err)
	}

	return Result{Module: module, Enabled: true, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Disable disables an Apache module (idempotent).
func Disable(module string) (Result, error) {
	start := time.Now()
	if module == "" {
		return Result{Error: "module must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("module must not be empty")
	}

	// Check if already disabled
	check, _ := Check(module)
	if !check.Enabled {
		return Result{Module: module, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	cmd := exec.Command("a2dismod", module)
	if out, err := cmd.CombinedOutput(); err != nil {
		return Result{Module: module, Error: string(out), DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("a2dismod: %s: %w", string(out), err)
	}

	return Result{Module: module, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}
