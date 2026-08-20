// Package include_vars provides Ansible include_vars module equivalent.
// Loads variables from YAML/JSON files into a shared variable store.
package include_vars

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// LoadResult is returned by Load.
type LoadResult struct {
	Changed  bool              `json:"changed"`
	File     string            `json:"file"`
	Count    int               `json:"count"`
	Variables map[string]string `json:"variables"`
	Error    string            `json:"error,omitempty"`
}

var (
	mu   sync.RWMutex
	vars = map[string]string{}
)

// Load reads variables from a YAML or JSON file and stores them.
func Load(filePath string) LoadResult {
	if filePath == "" {
		return LoadResult{Error: "file is required"}
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return LoadResult{Error: err.Error()}
	}

	parsed := map[string]interface{}{}
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return LoadResult{Error: "yaml parse: " + err.Error()}
		}
	case ".json":
		if err := json.Unmarshal(data, &parsed); err != nil {
			return LoadResult{Error: "json parse: " + err.Error()}
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			if err2 := json.Unmarshal(data, &parsed); err2 != nil {
				return LoadResult{Error: "unable to parse file as YAML or JSON"}
			}
		}
	}

	mu.Lock()
	defer mu.Unlock()
	flat := flattenMap("", parsed)
	for k, v := range flat {
		vars[k] = v
	}
	return LoadResult{Changed: true, File: filePath, Count: len(flat), Variables: flat}
}

// Get retrieves a variable value.
func Get(key string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	v, ok := vars[key]
	return v, ok
}

// GetAll returns all stored variables.
func GetAll() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	c := make(map[string]string, len(vars))
	for k, v := range vars {
		c[k] = v
	}
	return c
}

func flattenMap(prefix string, m map[string]interface{}) map[string]string {
	result := map[string]string{}
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			result[key] = val
		case map[string]interface{}:
			for fk, fv := range flattenMap(key, val) {
				result[fk] = fv
			}
		default:
			result[key] = serializeValue(v)
		}
	}
	return result
}

func serializeValue(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
