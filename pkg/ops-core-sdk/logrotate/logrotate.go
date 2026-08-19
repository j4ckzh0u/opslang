// Package logrotate provides log rotation configuration management.
package logrotate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents a logrotate configuration.
type Config struct {
	Name       string `json:"name"`
	Pattern    string `json:"pattern"`
	Frequency  string `json:"frequency"`
	Rotate     int    `json:"rotate"`
	Compress   bool   `json:"compress"`
	DelayCompress bool `json:"delay_compress"`
	MissingOK  bool   `json:"missing_ok"`
	NotIfEmpty bool   `json:"not_if_empty"`
	Create     bool   `json:"create"`
	CreateMode string `json:"create_mode,omitempty"`
	CreateOwner string `json:"create_owner,omitempty"`
	CreateGroup string `json:"create_group,omitempty"`
	PostRotate string `json:"post_rotate,omitempty"`
	FilePath   string `json:"file_path"`
}

// ListResult is returned by List.
type ListResult struct {
	Configs []Config `json:"configs"`
}

// GetResult is returned by Get.
type GetResult struct {
	Config Config `json:"config"`
}

// SetResult is returned by Set.
type SetResult struct {
	Changed bool   `json:"changed"`
	Name    string `json:"name"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	Name    string `json:"name"`
	Error   string `json:"error,omitempty"`
}

// List returns all logrotate configurations.
func List() (ListResult, error) {
	configs := []Config{}

	// Scan /etc/logrotate.d/
	pattern := "/etc/logrotate.d/*"
	files, err := filepath.Glob(pattern)
	if err != nil {
		return ListResult{Configs: configs}, fmt.Errorf("glob logrotate: %w", err)
	}

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil || info.IsDir() {
			continue
		}
		fileConfigs, err := parseLogrotateFile(f)
		if err != nil {
			continue
		}
		configs = append(configs, fileConfigs...)
	}

	return ListResult{Configs: configs}, nil
}

// Get returns a specific logrotate configuration by name.
func Get(name string) (GetResult, error) {
	configs, err := List()
	if err != nil {
		return GetResult{}, err
	}

	for _, c := range configs.Configs {
		if c.Name == name {
			return GetResult{Config: c}, nil
		}
	}

	return GetResult{}, fmt.Errorf("logrotate config %q not found", name)
}

// Set creates or updates a logrotate configuration.
func Set(name, pattern, frequency string, rotate int, compress bool, postRotate string) (SetResult, error) {
	if name == "" {
		return SetResult{Error: "name is required"}, fmt.Errorf("name is required")
	}
	if pattern == "" {
		return SetResult{Error: "pattern is required"}, fmt.Errorf("pattern is required")
	}

	filePath := fmt.Sprintf("/etc/logrotate.d/%s", name)

	// Check if already exists with same config
	if _, err := os.Stat(filePath); err == nil {
		existing, err := parseLogrotateFile(filePath)
		if err == nil && len(existing) > 0 {
			c := existing[0]
			if c.Pattern == pattern && c.Frequency == frequency && c.Rotate == rotate && c.Compress == compress {
				return SetResult{Changed: false, Name: name}, nil
			}
		}
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("%s {\n", pattern))
	content.WriteString(fmt.Sprintf("    %s\n", frequency))
	content.WriteString(fmt.Sprintf("    rotate %d\n", rotate))

	if compress {
		content.WriteString("    compress\n")
		content.WriteString("    delaycompress\n")
	}

	content.WriteString("    missingok\n")
	content.WriteString("    notifempty\n")
	content.WriteString("    create 0640 root root\n")

	if postRotate != "" {
		content.WriteString("    postrotate\n")
		content.WriteString(fmt.Sprintf("        %s\n", postRotate))
		content.WriteString("    endscript\n")
	}

	content.WriteString("}\n")

	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return SetResult{Error: err.Error(), Name: name}, fmt.Errorf("write logrotate config: %w", err)
	}

	return SetResult{Changed: true, Name: name}, nil
}

// Remove removes a logrotate configuration.
func Remove(name string) (RemoveResult, error) {
	if name == "" {
		return RemoveResult{Error: "name is required"}, fmt.Errorf("name is required")
	}

	filePath := fmt.Sprintf("/etc/logrotate.d/%s", name)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return RemoveResult{Changed: false, Name: name}, nil
	}

	if err := os.Remove(filePath); err != nil {
		return RemoveResult{Error: err.Error(), Name: name}, fmt.Errorf("remove logrotate config: %w", err)
	}

	return RemoveResult{Changed: true, Name: name}, nil
}

// parseLogrotateFile parses a logrotate configuration file.
func parseLogrotateFile(path string) ([]Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var configs []Config
	content := string(data)

	// Simple parser for single-block configs
	lines := strings.Split(content, "\n")
	var current *Config

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Opening brace with pattern
		if strings.HasSuffix(line, "{") {
			pattern := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			current = &Config{
				Name:    filepath.Base(path),
				Pattern: pattern,
				FilePath: path,
			}
			continue
		}

		// Closing brace
		if line == "}" {
			if current != nil {
				configs = append(configs, *current)
				current = nil
			}
			continue
		}

		if current == nil {
			continue
		}

		// Parse directives
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "daily":
			current.Frequency = "daily"
		case "weekly":
			current.Frequency = "weekly"
		case "monthly":
			current.Frequency = "monthly"
		case "yearly":
			current.Frequency = "yearly"
		case "rotate":
			if len(fields) >= 2 {
				fmt.Sscanf(fields[1], "%d", &current.Rotate)
			}
		case "compress":
			current.Compress = true
		case "delaycompress":
			current.DelayCompress = true
		case "missingok":
			current.MissingOK = true
		case "notifempty":
			current.NotIfEmpty = true
		case "create":
			current.Create = true
			if len(fields) >= 2 {
				current.CreateMode = fields[1]
			}
			if len(fields) >= 3 {
				current.CreateOwner = fields[2]
			}
			if len(fields) >= 4 {
				current.CreateGroup = fields[3]
			}
		}
	}

	return configs, nil
}
