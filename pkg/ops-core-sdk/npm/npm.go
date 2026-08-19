// Package npm provides Node.js npm package management operations.
package npm

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// Package represents an npm package.
type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ListResult represents the result of listing packages.
type ListResult struct {
	Packages []Package `json:"packages"`
}

// ActionResult represents the result of an npm action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// OutdatedPackage represents an outdated npm package.
type OutdatedPackage struct {
	Name    string `json:"name"`
	Current string `json:"current"`
	Wanted  string `json:"wanted"`
	Latest  string `json:"latest"`
}

// OutdatedResult represents the result of checking outdated packages.
type OutdatedResult struct {
	Packages []OutdatedPackage `json:"packages"`
}

// List returns installed npm packages.
func List(global bool) (*ListResult, error) {
	args := []string{"list", "--json", "--depth=0"}
	if global {
		args = append(args, "-g")
	}

	out, err := exec.Command("npm", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("npm list failed: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("failed to parse npm output: %w", err)
	}

	result := &ListResult{Packages: make([]Package, 0)}
	if deps, ok := data["dependencies"].(map[string]interface{}); ok {
		for name, info := range deps {
			if m, ok := info.(map[string]interface{}); ok {
				pkg := Package{Name: name}
				if v, ok := m["version"].(string); ok {
					pkg.Version = v
				}
				result.Packages = append(result.Packages, pkg)
			}
		}
	}

	return result, nil
}

// Install installs an npm package.
func Install(name string, global bool) (*ActionResult, error) {
	args := []string{"install", name}
	if global {
		args = append(args, "-g")
	}

	cmd := exec.Command("npm", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("npm install failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Installed %s", name),
	}, nil
}

// Uninstall uninstalls an npm package.
func Uninstall(name string, global bool) (*ActionResult, error) {
	args := []string{"uninstall", name}
	if global {
		args = append(args, "-g")
	}

	cmd := exec.Command("npm", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("npm uninstall failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Uninstalled %s", name),
	}, nil
}

// Outdated checks for outdated packages.
func Outdated(global bool) (*OutdatedResult, error) {
	args := []string{"outdated", "--json"}
	if global {
		args = append(args, "-g")
	}

	out, _ := exec.Command("npm", args...).CombinedOutput()

	var data map[string]interface{}
	if err := json.Unmarshal(out, &data); err != nil {
		return &OutdatedResult{Packages: make([]OutdatedPackage, 0)}, nil
	}

	result := &OutdatedResult{Packages: make([]OutdatedPackage, 0)}
	for name, info := range data {
		if m, ok := info.(map[string]interface{}); ok {
			pkg := OutdatedPackage{Name: name}
			if v, ok := m["current"].(string); ok {
				pkg.Current = v
			}
			if v, ok := m["wanted"].(string); ok {
				pkg.Wanted = v
			}
			if v, ok := m["latest"].(string); ok {
				pkg.Latest = v
			}
			result.Packages = append(result.Packages, pkg)
		}
	}

	return result, nil
}
