// Package set_fact provides Ansible set_fact module equivalent.
package set_fact

import "sync"

// SetFactResult is returned by Set.
type SetFactResult struct {
	Changed    bool              `json:"changed"`
	AnsibleFacts map[string]interface{} `json:"ansible_facts"`
}

var (
	mu    sync.RWMutex
	facts = map[string]interface{}{}
)

// Set registers one or more key-value facts.
func Set(keyValues map[string]interface{}) SetFactResult {
	mu.Lock()
	defer mu.Unlock()
	for k, v := range keyValues {
		facts[k] = v
	}
	return SetFactResult{Changed: true, AnsibleFacts: keyValues}
}

// Get retrieves a fact.
func Get(key string) (interface{}, bool) {
	mu.RLock()
	defer mu.RUnlock()
	v, ok := facts[key]
	return v, ok
}

// GetAll returns all facts.
func GetAll() map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]interface{}, len(facts))
	for k, v := range facts {
		out[k] = v
	}
	return out
}

// Clear removes all facts.
func Clear() SetFactResult {
	mu.Lock()
	defer mu.Unlock()
	facts = map[string]interface{}{}
	return SetFactResult{Changed: true, AnsibleFacts: map[string]interface{}{}}
}
