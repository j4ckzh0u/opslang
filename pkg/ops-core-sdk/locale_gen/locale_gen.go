package locale_gen

import (
	"fmt"
	"os/exec"
	"strings"
)

// GenerateResult represents the result of generating a locale.
type GenerateResult struct {
	Locale  string `json:"locale"`
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ListResult represents available locales.
type ListResult struct {
	Locales []string `json:"locales"`
	Count   int      `json:"count"`
	Error   string   `json:"error,omitempty"`
}

// RemoveResult represents the result of removing a locale.
type RemoveResult struct {
	Locale  string `json:"locale"`
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// Generate generates a locale.
func Generate(locale string) GenerateResult {
	if locale == "" {
		return GenerateResult{Error: "locale is required"}
	}

	cmd := exec.Command("locale-gen", locale)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return GenerateResult{
			Locale: locale,
			Status: "failed",
			Error:  fmt.Sprintf("locale-gen failed: %s: %s", err, string(out)),
		}
	}

	return GenerateResult{
		Locale:  locale,
		Status:  "success",
		Changed: true,
	}
}

// List lists available locales.
func List() ListResult {
	cmd := exec.Command("locale", "-a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ListResult{
			Error: fmt.Sprintf("locale -a failed: %s: %s", err, string(out)),
		}
	}

	var locales []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			locales = append(locales, line)
		}
	}

	return ListResult{
		Locales: locales,
		Count:   len(locales),
	}
}

// Remove removes a locale (by commenting it out in locale.gen).
func Remove(locale string) RemoveResult {
	if locale == "" {
		return RemoveResult{Error: "locale is required"}
	}

	// On Debian/Ubuntu, locales are managed via /etc/locale.gen
	// This is a simplified implementation
	cmd := exec.Command("sed", "-i", fmt.Sprintf("s/^%s /# %s /", locale, locale), "/etc/locale.gen")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return RemoveResult{
			Locale: locale,
			Status: "failed",
			Error:  fmt.Sprintf("failed to remove locale: %s: %s", err, string(out)),
		}
	}

	// Regenerate locales
	exec.Command("locale-gen").Run()

	return RemoveResult{
		Locale:  locale,
		Status:  "success",
		Changed: true,
	}
}
