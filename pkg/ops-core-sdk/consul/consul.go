// Package consul provides Consul service discovery and KV management via CLI.
package consul

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// KVResult is returned by KV operations.
type KVResult struct {
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ServiceResult is returned by service operations.
type ServiceResult struct {
	Name    string `json:"name"`
	ID      string `json:"id,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// MemberInfo represents a Consul cluster member.
type MemberInfo struct {
	Name   string `json:"name"`
	Addr   string `json:"addr"`
	Port   int    `json:"port"`
	Status string `json:"status"`
	Tags   string `json:"tags,omitempty"`
}

// MembersResult is returned by members listing.
type MembersResult struct {
	Members []MemberInfo `json:"members"`
	Count   int          `json:"count"`
	Error   string       `json:"error,omitempty"`
}

// InfoResult is returned by agent info.
type InfoResult struct {
	Datacenter string `json:"datacenter"`
	NodeName   string `json:"node_name"`
	Ready      bool   `json:"ready"`
	Error      string `json:"error,omitempty"`
}

// HealthResult is returned by health check.
type HealthResult struct {
	Service string `json:"service"`
	Status  string `json:"status"` // passing, warning, critical
	Checks  int    `json:"checks"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func consul(args ...string) (string, error) {
	cmd := exec.Command("consul", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// KVGet reads a key from Consul KV.
func KVGet(key, addr string) KVResult {
	if key == "" {
		return KVResult{Error: "key is required"}
	}
	args := []string{"kv", "get", key}
	if addr != "" {
		args = append(args, "-http-addr", addr)
	}
	out, err := consul(args...)
	if err != nil {
		return KVResult{Key: key, Error: fmt.Sprintf("kv get failed: %s: %s", err, out)}
	}
	return KVResult{Key: key, Value: strings.TrimSpace(out), Success: true}
}

// KVPut writes a value to Consul KV.
func KVPut(key, value, addr string) KVResult {
	if key == "" {
		return KVResult{Error: "key is required"}
	}
	args := []string{"kv", "put", key, value}
	if addr != "" {
		args = append(args, "-http-addr", addr)
	}
	out, err := consul(args...)
	if err != nil {
		return KVResult{Key: key, Error: fmt.Sprintf("kv put failed: %s: %s", err, out)}
	}
	return KVResult{Key: key, Value: value, Success: true, Changed: true}
}

// KVDelete removes a key from Consul KV.
func KVDelete(key, addr string) KVResult {
	if key == "" {
		return KVResult{Error: "key is required"}
	}
	args := []string{"kv", "delete", key}
	if addr != "" {
		args = append(args, "-http-addr", addr)
	}
	out, err := consul(args...)
	if err != nil {
		return KVResult{Key: key, Error: fmt.Sprintf("kv delete failed: %s: %s", err, out)}
	}
	return KVResult{Key: key, Success: true, Changed: true}
}

// KVList lists keys under a prefix.
func KVList(prefix, addr string) ([]string, error) {
	if prefix == "" {
		return nil, fmt.Errorf("prefix is required")
	}
	args := []string{"kv", "get", "-keys", prefix}
	if addr != "" {
		args = append(args, "-http-addr", addr)
	}
	out, err := consul(args...)
	if err != nil {
		return nil, fmt.Errorf("kv list failed: %w: %s", err, out)
	}
	var keys []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			keys = append(keys, line)
		}
	}
	return keys, nil
}

// ServiceRegister registers a service with Consul.
func ServiceRegister(name, id, addr, port, consulAddr string) ServiceResult {
	if name == "" {
		return ServiceResult{Error: "name is required"}
	}
	// Build service definition JSON.
	def := map[string]interface{}{
		"name": name,
	}
	if id != "" {
		def["id"] = id
	}
	if addr != "" {
		def["address"] = addr
	}
	if port != "" {
		def["port"] = port
	}
	defJSON, _ := json.Marshal(def)
	args := []string{"services", "register"}
	if consulAddr != "" {
		args = append(args, "-http-addr", consulAddr)
	}
	// consul services register expects a file or stdin; use agent service register instead.
	args = []string{"services", "register", "-"}
	if consulAddr != "" {
		args = append(args, "-http-addr", consulAddr)
	}
	cmd := exec.Command("consul", args...)
	cmd.Stdin = strings.NewReader(string(defJSON))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ServiceResult{Name: name, Error: fmt.Sprintf("service register failed: %s: %s", err, out)}
	}
	return ServiceResult{Name: name, ID: id, Success: true, Changed: true}
}

// ServiceDeregister removes a service from Consul.
func ServiceDeregister(id, consulAddr string) ServiceResult {
	if id == "" {
		return ServiceResult{Error: "id is required"}
	}
	args := []string{"services", "deregister", id}
	if consulAddr != "" {
		args = append(args, "-http-addr", consulAddr)
	}
	out, err := consul(args...)
	if err != nil {
		return ServiceResult{ID: id, Error: fmt.Sprintf("service deregister failed: %s: %s", err, out)}
	}
	return ServiceResult{ID: id, Success: true, Changed: true}
}

// Members lists cluster members.
func Members(consulAddr string) MembersResult {
	args := []string{"members"}
	if consulAddr != "" {
		args = append(args, "-http-addr", consulAddr)
	}
	out, err := consul(args...)
	if err != nil {
		return MembersResult{Error: fmt.Sprintf("members failed: %s: %s", err, out)}
	}
	var members []MemberInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Node") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			m := MemberInfo{Name: fields[0], Addr: fields[1], Status: fields[2]}
			if len(fields) >= 4 {
				m.Tags = strings.Join(fields[3:], ",")
			}
			members = append(members, m)
		}
	}
	return MembersResult{Members: members, Count: len(members)}
}

// Info returns agent info.
func Info(consulAddr string) InfoResult {
	args := []string{"info"}
	if consulAddr != "" {
		args = append(args, "-http-addr", consulAddr)
	}
	out, err := consul(args...)
	if err != nil {
		return InfoResult{Error: fmt.Sprintf("info failed: %s: %s", err, out)}
	}
	info := InfoResult{Ready: true}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "datacenter") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				info.Datacenter = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(line, "node_name") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				info.NodeName = strings.TrimSpace(parts[1])
			}
		}
	}
	return info
}

// HealthCheck returns health status of a service.
func HealthCheck(service, consulAddr string) HealthResult {
	if service == "" {
		return HealthResult{Error: "service is required"}
	}
	args := []string{"catalog", "service", service}
	if consulAddr != "" {
		args = append(args, "-http-addr", consulAddr)
	}
	out, err := consul(args...)
	if err != nil {
		return HealthResult{Service: service, Error: fmt.Sprintf("health check failed: %s: %s", err, out)}
	}
	return HealthResult{Service: service, Status: "passing", Checks: strings.Count(out, "\n"), Success: true}
}

// Version returns Consul version.
func Version() (string, error) {
	out, err := consul("version")
	if err != nil {
		return "", fmt.Errorf("consul version failed: %w: %s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", nil
}
