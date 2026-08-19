// Package apache2 provides Apache2 web server management operations.
package apache2

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ActionResult represents the result of an Apache2 action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// ConfigTestResult represents the result of config test.
type ConfigTestResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// SitesListResult represents the result of listing sites.
type SitesListResult struct {
	Enabled   []string `json:"enabled"`
	Available []string `json:"available"`
}

// ModulesListResult represents the result of listing modules.
type ModulesListResult struct {
	Enabled []string `json:"enabled"`
}

// detectCmd finds the apache2 control binary.
func detectCmd() string {
	for _, name := range []string{"apache2ctl", "apachectl", "httpd"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return "apache2ctl"
}

// ConfigTest tests the Apache2 configuration.
func ConfigTest() (*ConfigTestResult, error) {
	cmd := exec.Command(detectCmd(), "-t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ConfigTestResult{
			Valid:   false,
			Message: strings.TrimSpace(string(out)),
		}, nil
	}
	return &ConfigTestResult{
		Valid:   true,
		Message: "Syntax OK",
	}, nil
}

// Reload gracefully reloads Apache2 configuration.
func Reload() (*ActionResult, error) {
	cmd := exec.Command(detectCmd(), "-k", "graceful")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("apache2 reload failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: "Apache2 reloaded"}, nil
}

// SitesList returns enabled and available sites.
func SitesList() (*SitesListResult, error) {
	result := &SitesListResult{
		Enabled:   make([]string, 0),
		Available: make([]string, 0),
	}

	for _, dir := range []string{"/etc/apache2/sites-enabled", "/etc/apache2/sites-available"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if strings.HasSuffix(name, ".conf") {
				name = strings.TrimSuffix(name, ".conf")
			}
			if filepath.Base(dir) == "sites-enabled" {
				result.Enabled = append(result.Enabled, name)
			} else {
				result.Available = append(result.Available, name)
			}
		}
	}
	return result, nil
}

// SiteEnable enables an Apache2 site.
func SiteEnable(name string) (*ActionResult, error) {
	cmd := exec.Command("a2ensite", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("a2ensite failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Enabled site %s", name)}, nil
}

// SiteDisable disables an Apache2 site.
func SiteDisable(name string) (*ActionResult, error) {
	cmd := exec.Command("a2dissite", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("a2dissite failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Disabled site %s", name)}, nil
}

// ModulesList returns enabled modules.
func ModulesList() (*ModulesListResult, error) {
	cmd := exec.Command(detectCmd(), "-M")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("apache2 module list failed: %w (output: %s)", err, string(out))
	}

	result := &ModulesListResult{Enabled: make([]string, 0)}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "_module") {
			name := strings.TrimSuffix(line, "_module")
			if idx := strings.Index(name, " ("); idx > 0 {
				name = name[:idx]
			}
			result.Enabled = append(result.Enabled, name)
		}
	}
	return result, nil
}

// ModuleEnable enables an Apache2 module.
func ModuleEnable(name string) (*ActionResult, error) {
	cmd := exec.Command("a2enmod", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("a2enmod failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Enabled module %s", name)}, nil
}

// ModuleDisable disables an Apache2 module.
func ModuleDisable(name string) (*ActionResult, error) {
	cmd := exec.Command("a2dismod", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("a2dismod failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Disabled module %s", name)}, nil
}
