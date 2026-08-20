// Package add_host provides Ansible add_host module equivalent.
// Add a host to the in-memory inventory during playbook execution.
package add_host

import "sync"

// AddHostResult is returned by Add.
type AddHostResult struct {
	Name    string   `json:"name"`
	Groups  []string `json:"groups,omitempty"`
	Changed bool     `json:"changed"`
	Error   string   `json:"error,omitempty"`
}

var (
	mu    sync.RWMutex
	hosts = map[string]map[string]string{} // host -> vars
	groups = map[string][]string{}        // group -> hosts
)

// Add adds a host to inventory with optional groups and variables.
func Add(name string, groupsToAdd []string, vars map[string]string) AddHostResult {
	if name == "" {
		return AddHostResult{Error: "name is required"}
	}
	mu.Lock()
	defer mu.Unlock()

	if _, ok := hosts[name]; !ok {
		hosts[name] = make(map[string]string)
	}
	for k, v := range vars {
		hosts[name][k] = v
	}

	for _, g := range groupsToAdd {
		found := false
		for _, h := range groups[g] {
			if h == name {
				found = true
				break
			}
		}
		if !found {
			groups[g] = append(groups[g], name)
		}
	}

	return AddHostResult{Name: name, Groups: groupsToAdd, Changed: true}
}

// GetHost returns host variables.
func GetHost(name string) (map[string]string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	v, ok := hosts[name]
	return v, ok
}

// GetGroup returns hosts in a group.
func GetGroup(group string) []string {
	mu.RLock()
	defer mu.RUnlock()
	return groups[group]
}

// ListHosts returns all host names.
func ListHosts() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(hosts))
	for n := range hosts {
		names = append(names, n)
	}
	return names
}

// ListGroups returns all group names.
func ListGroups() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(groups))
	for n := range groups {
		names = append(names, n)
	}
	return names
}
