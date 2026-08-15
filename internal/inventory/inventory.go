// Package inventory provides parsing for YAML-based host inventory files.
package inventory

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Host represents a single host in the inventory.
type Host struct {
	Name     string `yaml:"name" json:"name"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"-"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
	Group    string `yaml:"group" json:"group"`
	Tags     map[string]string `yaml:"tags" json:"tags"`
}

// Inventory represents the full inventory file structure.
type Inventory struct {
	Hosts []Host `yaml:"hosts" json:"hosts"`
}

// ParseFile reads and parses a YAML inventory file.
func ParseFile(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read inventory file %s: %w", path, err)
	}

	return Parse(data)
}

// Parse parses YAML inventory data.
func Parse(data []byte) (*Inventory, error) {
	var inv Inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("failed to parse inventory: %w", err)
	}

	if err := validate(&inv); err != nil {
		return nil, fmt.Errorf("invalid inventory: %w", err)
	}

	return &inv, nil
}

// validate checks the inventory for required fields and applies defaults.
func validate(inv *Inventory) error {
	for i := range inv.Hosts {
		h := &inv.Hosts[i]

		if h.Host == "" {
			if h.Name != "" {
				h.Host = h.Name
			} else {
				return fmt.Errorf("host %d: either 'name' or 'host' is required", i)
			}
		}

		if h.Name == "" {
			h.Name = h.Host
		}

		if h.Port == 0 {
			h.Port = 22
		}

		if h.User == "" {
			h.User = "root"
		}
	}

	return nil
}

// GetHostsByGroup returns hosts that belong to the specified group.
func (inv *Inventory) GetHostsByGroup(group string) []Host {
	var result []Host
	for _, h := range inv.Hosts {
		if h.Group == group {
			result = append(result, h)
		}
	}
	return result
}

// GetHostsByTag returns hosts that have the specified tag key-value pair.
func (inv *Inventory) GetHostsByTag(key, value string) []Host {
	var result []Host
	for _, h := range inv.Hosts {
		if h.Tags != nil && h.Tags[key] == value {
			result = append(result, h)
		}
	}
	return result
}

// GetHostByName returns a host by its name.
func (inv *Inventory) GetHostByName(name string) (*Host, error) {
	for _, h := range inv.Hosts {
		if h.Name == name {
			return &h, nil
		}
	}
	return nil, fmt.Errorf("host %q not found", name)
}

// AllHostNames returns all host names in the inventory.
func (inv *Inventory) AllHostNames() []string {
	names := make([]string, len(inv.Hosts))
	for i, h := range inv.Hosts {
		names[i] = h.Name
	}
	return names
}
