// Package apt_repo provides APT repository management operations.
package apt_repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoInfo represents an APT repository source.
type RepoInfo struct {
	Type    string `json:"type"`
	URI     string `json:"uri"`
	Dist    string `json:"dist"`
	Comps   string `json:"components"`
	Enabled bool   `json:"enabled"`
	File    string `json:"file"`
	Line    string `json:"line"`
}

// ListResult is returned by List.
type ListResult struct {
	Repos []RepoInfo `json:"repos"`
}

// ExistsResult is returned by Exists.
type ExistsResult struct {
	Exists bool   `json:"exists"`
	URI    string `json:"uri,omitempty"`
}

// AddResult is returned by Add.
type AddResult struct {
	Changed bool   `json:"changed"`
	URI     string `json:"uri"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	URI     string `json:"uri"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult is returned by Update.
type UpdateResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// List returns all configured APT repositories.
func List() (ListResult, error) {
	repos := []RepoInfo{}

	// Scan /etc/apt/sources.list
	if info, err := os.Stat("/etc/apt/sources.list"); err == nil && !info.IsDir() {
		fileRepos, err := parseSourcesFile("/etc/apt/sources.list")
		if err == nil {
			repos = append(repos, fileRepos...)
		}
	}

	// Scan /etc/apt/sources.list.d/*.list
	pattern := "/etc/apt/sources.list.d/*.list"
	files, err := filepath.Glob(pattern)
	if err != nil {
		return ListResult{Repos: repos}, fmt.Errorf("glob sources: %w", err)
	}

	for _, f := range files {
		fileRepos, err := parseSourcesFile(f)
		if err != nil {
			continue
		}
		repos = append(repos, fileRepos...)
	}

	return ListResult{Repos: repos}, nil
}

// Exists checks if an APT repository with the given URI exists.
func Exists(uri string) (ExistsResult, error) {
	repos, err := List()
	if err != nil {
		return ExistsResult{}, err
	}

	for _, r := range repos.Repos {
		if strings.Contains(r.URI, uri) || strings.Contains(uri, r.URI) {
			return ExistsResult{Exists: true, URI: r.URI}, nil
		}
	}
	return ExistsResult{Exists: false, URI: uri}, nil
}

// Add adds an APT repository source.
func Add(uri, dist, components string) (AddResult, error) {
	if uri == "" {
		return AddResult{Error: "URI is required"}, fmt.Errorf("URI is required")
	}

	// Check if already exists
	existing, _ := Exists(uri)
	if existing.Exists {
		return AddResult{Changed: false, URI: uri}, nil
	}

	// Create repo file
	safeName := strings.ReplaceAll(uri, "/", "_")
	safeName = strings.ReplaceAll(safeName, ":", "_")
	filePath := fmt.Sprintf("/etc/apt/sources.list.d/%s.list", safeName)

	line := fmt.Sprintf("deb %s %s %s\n", uri, dist, components)
	if err := os.WriteFile(filePath, []byte(line), 0644); err != nil {
		return AddResult{Error: err.Error(), URI: uri}, fmt.Errorf("write repo file: %w", err)
	}

	return AddResult{Changed: true, URI: uri}, nil
}

// Remove removes an APT repository by URI.
func Remove(uri string) (RemoveResult, error) {
	if uri == "" {
		return RemoveResult{Error: "URI is required"}, fmt.Errorf("URI is required")
	}

	// Find the file containing this URI
	repos, err := List()
	if err != nil {
		return RemoveResult{Error: err.Error(), URI: uri}, err
	}

	found := false
	for _, r := range repos.Repos {
		if strings.Contains(r.URI, uri) {
			found = true
			// Remove the line from the file
			if err := removeLineFromFile(r.File, r.Line); err != nil {
				return RemoveResult{Error: err.Error(), URI: uri}, err
			}
		}
	}

	if !found {
		return RemoveResult{Changed: false, URI: uri}, nil
	}

	return RemoveResult{Changed: true, URI: uri}, nil
}

// Update runs apt-get update to refresh package lists.
func Update() (UpdateResult, error) {
	if _, err := exec.LookPath("apt-get"); err != nil {
		return UpdateResult{Error: "apt-get not found"}, fmt.Errorf("apt-get not found")
	}

	cmd := exec.Command("apt-get", "update")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return UpdateResult{Error: stderr.String()}, fmt.Errorf("apt-get update: %w", err)
	}

	return UpdateResult{Changed: true}, nil
}

// parseSourcesFile parses a sources.list file.
func parseSourcesFile(path string) ([]RepoInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var repos []RepoInfo
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		if fields[0] != "deb" && fields[0] != "deb-src" {
			continue
		}

		repo := RepoInfo{
			Type:    fields[0],
			URI:     fields[1],
			Dist:    fields[2],
			Comps:   strings.Join(fields[3:], " "),
			Enabled: !strings.HasPrefix(line, "#"),
			File:    path,
			Line:    line,
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// removeLineFromFile removes a specific line from a file.
func removeLineFromFile(path, targetLine string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != strings.TrimSpace(targetLine) {
			lines = append(lines, line)
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
