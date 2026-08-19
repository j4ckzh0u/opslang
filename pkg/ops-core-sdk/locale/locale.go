// Package locale provides system locale management operations.
package locale

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// LocaleInfo represents locale information.
type LocaleInfo struct {
	Lang     string `json:"lang"`
	Language string `json:"language,omitempty"`
	LcAll    string `json:"lc_all,omitempty"`
}

// AvailableResult is returned by Available.
type AvailableResult struct {
	Locales []string `json:"locales"`
}

// SetResult is returned by Set.
type SetResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// Get returns the current system locale settings.
func Get() (LocaleInfo, error) {
	cmd := exec.Command("locale")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return LocaleInfo{}, fmt.Errorf("locale command failed: %w", err)
	}

	info := LocaleInfo{}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.HasPrefix(line, "LANG=") {
			info.Lang = strings.Trim(strings.TrimPrefix(line, "LANG="), `"`)
		} else if strings.HasPrefix(line, "LANGUAGE=") {
			info.Language = strings.Trim(strings.TrimPrefix(line, "LANGUAGE="), `"`)
		} else if strings.HasPrefix(line, "LC_ALL=") {
			info.LcAll = strings.Trim(strings.TrimPrefix(line, "LC_ALL="), `"`)
		}
	}

	return info, nil
}

// Available returns all available locales on the system.
func Available() (AvailableResult, error) {
	cmd := exec.Command("locale", "-a")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return AvailableResult{}, fmt.Errorf("locale -a failed: %w", err)
	}

	var locales []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			locales = append(locales, line)
		}
	}

	return AvailableResult{Locales: locales}, nil
}

// Set sets the system locale.
func Set(locale string) (SetResult, error) {
	if locale == "" {
		return SetResult{Error: "locale is required"}, fmt.Errorf("locale is required")
	}

	// Check current locale
	current, err := Get()
	if err != nil {
		return SetResult{Error: err.Error()}, err
	}

	if current.Lang == locale && current.LcAll == locale {
		return SetResult{Changed: false}, nil
	}

	// Try localectl first (systemd-based systems)
	if _, err := exec.LookPath("localectl"); err == nil {
		cmd := exec.Command("localectl", "set-locale", fmt.Sprintf("LANG=%s", locale))
		if err := cmd.Run(); err == nil {
			return SetResult{Changed: true}, nil
		}
	}

	// Fallback to updating /etc/default/locale or /etc/locale.conf
	localeFile := "/etc/default/locale"
	if _, err := os.Stat("/etc/locale.conf"); err == nil {
		localeFile = "/etc/locale.conf"
	}

	content, err := os.ReadFile(localeFile)
	if err != nil {
		// File doesn't exist, create it
		content = []byte{}
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "LANG=") || strings.HasPrefix(line, "LC_ALL=") {
			continue
		}
		newLines = append(newLines, line)
	}

	newLines = append(newLines, fmt.Sprintf("LANG=%s", locale))
	newLines = append(newLines, fmt.Sprintf("LC_ALL=%s", locale))

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(localeFile, []byte(newContent), 0644); err != nil {
		return SetResult{Error: err.Error()}, fmt.Errorf("write %s: %w", localeFile, err)
	}

	return SetResult{Changed: true}, nil
}
