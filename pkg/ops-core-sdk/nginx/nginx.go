// Package nginx provides Nginx web server management operations.
package nginx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ActionResult represents the result of an nginx action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// Site represents an nginx site.
type Site struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// SitesResult represents the result of listing sites.
type SitesResult struct {
	Sites []Site `json:"sites"`
}

// ConfigTestResult represents the result of testing config.
type ConfigTestResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// ConfigTest tests nginx configuration.
func ConfigTest() (*ConfigTestResult, error) {
	out, err := exec.Command("nginx", "-t").CombinedOutput()

	valid := err == nil
	message := string(out)

	return &ConfigTestResult{
		Valid:   valid,
		Message: message,
	}, nil
}

// Reload reloads nginx configuration.
func Reload() (*ActionResult, error) {
	out, err := exec.Command("nginx", "-s", "reload").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nginx reload failed: %w (output: %s)", err, string(out))
	}

	return &ActionResult{
		Changed: true,
		Message: "Nginx reloaded",
	}, nil
}

// SitesList lists all nginx sites.
func SitesList() (*SitesResult, error) {
	sites := make([]Site, 0)

	// Check sites-available
	availableDir := "/etc/nginx/sites-available"
	enabledDir := "/etc/nginx/sites-enabled"

	if _, err := os.Stat(availableDir); os.IsNotExist(err) {
		return &SitesResult{Sites: sites}, nil
	}

	files, err := os.ReadDir(availableDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sites-available: %w", err)
	}

	enabledSites := make(map[string]bool)
	if files, err := os.ReadDir(enabledDir); err == nil {
		for _, f := range files {
			enabledSites[f.Name()] = true
		}
	}

	for _, f := range files {
		if !f.IsDir() {
			site := Site{
				Name:    f.Name(),
				Enabled: enabledSites[f.Name()],
			}
			sites = append(sites, site)
		}
	}

	return &SitesResult{Sites: sites}, nil
}

// SiteEnable enables an nginx site.
func SiteEnable(name string) (*ActionResult, error) {
	available := filepath.Join("/etc/nginx/sites-available", name)
	enabled := filepath.Join("/etc/nginx/sites-enabled", name)

	if _, err := os.Stat(available); os.IsNotExist(err) {
		return nil, fmt.Errorf("site %s does not exist in sites-available", name)
	}

	// Check if already enabled
	if _, err := os.Stat(enabled); err == nil {
		return &ActionResult{Changed: false, Message: fmt.Sprintf("Site %s already enabled", name)}, nil
	}

	if err := os.Symlink(available, enabled); err != nil {
		return nil, fmt.Errorf("failed to enable site: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Enabled site %s", name),
	}, nil
}

// SiteDisable disables an nginx site.
func SiteDisable(name string) (*ActionResult, error) {
	enabled := filepath.Join("/etc/nginx/sites-enabled", name)

	if _, err := os.Stat(enabled); os.IsNotExist(err) {
		return &ActionResult{Changed: false, Message: fmt.Sprintf("Site %s already disabled", name)}, nil
	}

	if err := os.Remove(enabled); err != nil {
		return nil, fmt.Errorf("failed to disable site: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Disabled site %s", name),
	}, nil
}
