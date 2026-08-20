// Package group_by provides Ansible group_by module equivalent.
// Creates ad-hoc inventory groups based on a key expression.
package group_by

import (
	"sync"
)

// GroupByResult is returned by GroupBy.
type GroupByResult struct {
	Changed bool     `json:"changed"`
	Groups  []string `json:"groups"`
	Key     string   `json:"key"`
	Error   string   `json:"error,omitempty"`
}

var (
	mu     sync.RWMutex
	groups = map[string][]string{} // group name -> host list
)

// GroupBy creates or adds hosts to an ad-hoc group.
func GroupBy(hosts []string, key string) GroupByResult {
	if key == "" {
		return GroupByResult{Error: "key is required"}
	}
	if len(hosts) == 0 {
		return GroupByResult{Key: key, Groups: []string{key}}
	}
	mu.Lock()
	defer mu.Unlock()

	existing := groups[key]
	seen := map[string]bool{}
	for _, h := range existing {
		seen[h] = true
	}
	for _, h := range hosts {
		if !seen[h] {
			existing = append(existing, h)
			seen[h] = true
		}
	}
	groups[key] = existing
	return GroupByResult{Changed: true, Key: key, Groups: []string{key}}
}

// GetHosts returns the hosts in a group.
func GetHosts(group string) []string {
	mu.RLock()
	defer mu.RUnlock()
	h := groups[group]
	if h == nil {
		return []string{}
	}
	c := make([]string, len(h))
	copy(c, h)
	return c
}

// ListGroups returns all group names.
func ListGroups() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(groups))
	for k := range groups {
		names = append(names, k)
	}
	return names
}

// Clear removes all groups.
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	groups = map[string][]string{}
}
