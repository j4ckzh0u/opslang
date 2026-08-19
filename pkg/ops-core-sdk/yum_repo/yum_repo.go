// Package yum_repo provides YUM/DNF repository management operations.
package yum_repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RepoInfo represents a YUM repository configuration.
type RepoInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	Enabled  bool   `json:"enabled"`
	GPGCheck bool   `json:"gpg_check"`
	GPGKey   string `json:"gpg_key,omitempty"`
	FilePath string `json:"file_path"`
}

// ListResult is returned by List.
type ListResult struct {
	Repos []RepoInfo `json:"repos"`
}

// ExistsResult is returned by Exists.
type ExistsResult struct {
	Exists bool   `json:"exists"`
	ID     string `json:"id,omitempty"`
}

// AddResult is returned by Add.
type AddResult struct {
	Changed bool   `json:"changed"`
	ID      string `json:"id"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	ID      string `json:"id"`
	Error   string `json:"error,omitempty"`
}

// List returns all configured YUM repositories.
func List() (ListResult, error) {
	repos := []RepoInfo{}

	// Scan /etc/yum.repos.d/
	pattern := "/etc/yum.repos.d/*.repo"
	files, err := filepath.Glob(pattern)
	if err != nil {
		return ListResult{Repos: repos}, fmt.Errorf("glob repos: %w", err)
	}

	for _, f := range files {
		fileRepos, err := parseRepoFile(f)
		if err != nil {
			continue
		}
		repos = append(repos, fileRepos...)
	}

	return ListResult{Repos: repos}, nil
}

// Exists checks if a YUM repository with the given ID exists.
func Exists(id string) (ExistsResult, error) {
	repos, err := List()
	if err != nil {
		return ExistsResult{}, err
	}

	for _, r := range repos.Repos {
		if r.ID == id {
			return ExistsResult{Exists: true, ID: id}, nil
		}
	}
	return ExistsResult{Exists: false, ID: id}, nil
}

// Add adds or updates a YUM repository.
func Add(id, name, baseURL string, gpgCheck bool, gpgKey string) (AddResult, error) {
	if id == "" {
		return AddResult{Error: "repo ID is required"}, fmt.Errorf("repo ID is required")
	}

	// Check if already exists
	existing, _ := Exists(id)
	if existing.Exists {
		return AddResult{Changed: false, ID: id}, nil
	}

	// Create repo file
	filePath := fmt.Sprintf("/etc/yum.repos.d/%s.repo", id)

	enabled := "1"
	if !gpgCheck {
		enabled = "0"
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("[%s]\n", id))
	content.WriteString(fmt.Sprintf("name=%s\n", name))
	content.WriteString(fmt.Sprintf("baseurl=%s\n", baseURL))
	content.WriteString(fmt.Sprintf("enabled=%s\n", enabled))
	content.WriteString(fmt.Sprintf("gpgcheck=%v\n", boolToInt(gpgCheck)))
	if gpgKey != "" {
		content.WriteString(fmt.Sprintf("gpgkey=%s\n", gpgKey))
	}

	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return AddResult{Error: err.Error(), ID: id}, fmt.Errorf("write repo file: %w", err)
	}

	return AddResult{Changed: true, ID: id}, nil
}

// Remove removes a YUM repository by ID.
func Remove(id string) (RemoveResult, error) {
	if id == "" {
		return RemoveResult{Error: "repo ID is required"}, fmt.Errorf("repo ID is required")
	}

	filePath := fmt.Sprintf("/etc/yum.repos.d/%s.repo", id)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return RemoveResult{Changed: false, ID: id}, nil
	}

	if err := os.Remove(filePath); err != nil {
		return RemoveResult{Error: err.Error(), ID: id}, fmt.Errorf("remove repo file: %w", err)
	}

	return RemoveResult{Changed: true, ID: id}, nil
}

// parseRepoFile parses a .repo file and returns all repo entries.
func parseRepoFile(path string) ([]RepoInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var repos []RepoInfo
	var current *RepoInfo

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// New repo section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if current != nil {
				current.FilePath = path
				repos = append(repos, *current)
			}
			id := strings.Trim(line, "[]")
			current = &RepoInfo{ID: id}
			continue
		}

		if current == nil {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			current.Name = value
		case "baseurl":
			current.BaseURL = value
		case "enabled":
			current.Enabled = value == "1"
		case "gpgcheck":
			current.GPGCheck = value == "1"
		case "gpgkey":
			current.GPGKey = value
		}
	}

	if current != nil {
		current.FilePath = path
		repos = append(repos, *current)
	}

	return repos, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
