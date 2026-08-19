// Package firewalld_ipset manages firewalld IP sets.
// Equivalent to ansible.posix.firewalld_ipset module.
package firewalld_ipset

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Status string   `json:"status"`
	Sets   []string `json:"sets"`
	Count  int      `json:"count"`
	Error  string   `json:"error,omitempty"`
}

// InfoResult is returned by Info.
type InfoResult struct {
	Status  string   `json:"status"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Entries []string `json:"entries"`
	Count   int      `json:"count"`
	Error   string   `json:"error,omitempty"`
}

// Create creates a new IP set.
// setType: "hash:ip", "hash:net", "hash:mac", etc.
func Create(name, setType string) Result {
	if name == "" || setType == "" {
		return Result{Status: "failed", Error: "name and type are required"}
	}

	cmd := exec.Command("firewall-cmd", "--permanent", "--new-ipset="+name, "--type="+setType)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Name: name, Type: setType,
			Error: fmt.Sprintf("firewall-cmd: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	exec.Command("firewall-cmd", "--reload").Run()

	return Result{Status: "success", Changed: true, Name: name, Type: setType}
}

// Delete deletes an IP set.
func Delete(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}

	cmd := exec.Command("firewall-cmd", "--permanent", "--delete-ipset="+name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Name: name,
			Error: fmt.Sprintf("firewall-cmd: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	exec.Command("firewall-cmd", "--reload").Run()

	return Result{Status: "success", Changed: true, Name: name}
}

// AddEntry adds an entry to an IP set.
func AddEntry(name, entry string) Result {
	if name == "" || entry == "" {
		return Result{Status: "failed", Error: "name and entry are required"}
	}

	cmd := exec.Command("firewall-cmd", "--permanent", "--ipset="+name, "--add-entry="+entry)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Name: name,
			Error: fmt.Sprintf("firewall-cmd: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	exec.Command("firewall-cmd", "--reload").Run()

	return Result{Status: "success", Changed: true, Name: name}
}

// RemoveEntry removes an entry from an IP set.
func RemoveEntry(name, entry string) Result {
	if name == "" || entry == "" {
		return Result{Status: "failed", Error: "name and entry are required"}
	}

	cmd := exec.Command("firewall-cmd", "--permanent", "--ipset="+name, "--remove-entry="+entry)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Name: name,
			Error: fmt.Sprintf("firewall-cmd: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	exec.Command("firewall-cmd", "--reload").Run()

	return Result{Status: "success", Changed: true, Name: name}
}

// List lists all IP sets.
func List() ListResult {
	cmd := exec.Command("firewall-cmd", "--permanent", "--get-ipsets")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed",
			Error: fmt.Sprintf("firewall-cmd: %v", err)}
	}

	var sets []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			sets = append(sets, line)
		}
	}

	return ListResult{
		Status: "success",
		Sets:   sets,
		Count:  len(sets),
	}
}

// Info returns information about an IP set.
func Info(name string) InfoResult {
	if name == "" {
		return InfoResult{Status: "failed", Error: "name is required"}
	}

	// Get entries
	cmd := exec.Command("firewall-cmd", "--permanent", "--ipset="+name, "--get-entries")
	out, err := cmd.Output()
	if err != nil {
		return InfoResult{Status: "failed", Name: name,
			Error: fmt.Sprintf("firewall-cmd: %v", err)}
	}

	var entries []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			entries = append(entries, line)
		}
	}

	return InfoResult{
		Status:  "success",
		Name:    name,
		Entries: entries,
		Count:   len(entries),
	}
}
