// Package set_stats provides Ansible set_stats module equivalent.
// Allows setting custom stats that can be used by Ansible Tower/AWX.
package set_stats

import (
	"sync"
)

// StatsResult is returned by Set.
type StatsResult struct {
	Changed bool              `json:"changed"`
	Data    map[string]string `json:"data"`
	Error   string            `json:"error,omitempty"`
}

var (
	mu    sync.RWMutex
	stats = map[string]string{}
)

// Set stores custom stats.
func Set(data map[string]string) StatsResult {
	if data == nil || len(data) == 0 {
		return StatsResult{Error: "data is required"}
	}
	mu.Lock()
	defer mu.Unlock()
	for k, v := range data {
		stats[k] = v
	}
	return StatsResult{Changed: true, Data: copyStats()}
}

// Get retrieves a single stat value.
func Get(key string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	v, ok := stats[key]
	return v, ok
}

// GetAll returns all stored stats.
func GetAll() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	return copyStats()
}

// Clear removes all stored stats.
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	stats = map[string]string{}
}

func copyStats() map[string]string {
	c := make(map[string]string, len(stats))
	for k, v := range stats {
		c[k] = v
	}
	return c
}
