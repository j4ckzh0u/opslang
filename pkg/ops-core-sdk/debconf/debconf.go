// Package debconf manages Debian package configuration selections.
// Equivalent to ansible.builtin.debconf module.
package debconf

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SetResult is returned by Set.
type SetResult struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Package string `json:"package"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Error   string `json:"error,omitempty"`
}

// GetResult is returned by Get.
type GetResult struct {
	Status  string `json:"status"`
	Package string `json:"package"`
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"` // select, boolean, string, etc.
	Found   bool   `json:"found"`
	Error   string `json:"error,omitempty"`
}

// ListItem represents one debconf selection.
type ListItem struct {
	Package string `json:"package"`
	Key     string `json:"key"`
	Type    string `json:"type"`
	Value   string `json:"value"`
}

// ListResult is returned by List.
type ListResult struct {
	Status  string     `json:"status"`
	Package string     `json:"package"`
	Items   []ListItem `json:"items"`
	Count   int        `json:"count"`
	Error   string     `json:"error,omitempty"`
}

// Set sets a debconf value using debconf-set-selections.
func Set(pkg, name, vtype, value string) SetResult {
	if pkg == "" || name == "" || value == "" {
		return SetResult{Status: "failed", Error: "package, name, and value are required"}
	}
	if vtype == "" {
		vtype = "string"
	}

	// Format: package question type value
	line := fmt.Sprintf("%s %s %s %s", pkg, name, vtype, value)
	cmd := exec.Command("debconf-set-selections")
	cmd.Stdin = strings.NewReader(line + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return SetResult{Status: "failed", Package: pkg, Key: name,
			Error: fmt.Sprintf("debconf-set-selections: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return SetResult{Status: "success", Changed: true, Package: pkg, Key: name, Value: value}
}

// Get retrieves a debconf value using debconf-get-selections.
func Get(pkg, name string) GetResult {
	if pkg == "" || name == "" {
		return GetResult{Status: "failed", Error: "package and name are required"}
	}

	cmd := exec.Command("debconf-get-selections")
	out, err := cmd.Output()
	if err != nil {
		// debconf-get-selections may not be installed
		if _, lookErr := exec.LookPath("debconf-get-selections"); lookErr != nil {
			return GetResult{Status: "failed", Package: pkg, Key: name,
				Error: "debconf-get-selections not found; install debconf-utils"}
		}
		return GetResult{Status: "failed", Package: pkg, Key: name,
			Error: fmt.Sprintf("debconf-get-selections: %v", err)}
	}

	// Parse: package question type value
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if fields[0] == pkg && fields[1] == name {
			value := strings.Join(fields[3:], " ")
			return GetResult{
				Status:  "success",
				Package: pkg,
				Key:     name,
				Type:    fields[2],
				Value:   value,
				Found:   true,
			}
		}
	}
	return GetResult{Status: "success", Package: pkg, Key: name, Found: false}
}

// List lists all debconf selections for a package.
func List(pkg string) ListResult {
	if pkg == "" {
		return ListResult{Status: "failed", Error: "package is required"}
	}

	cmd := exec.Command("debconf-get-selections")
	out, err := cmd.Output()
	if err != nil {
		if _, lookErr := exec.LookPath("debconf-get-selections"); lookErr != nil {
			return ListResult{Status: "failed", Package: pkg,
				Error: "debconf-get-selections not found; install debconf-utils"}
		}
		return ListResult{Status: "failed", Package: pkg,
			Error: fmt.Sprintf("debconf-get-selections: %v", err)}
	}

	var items []ListItem
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] != pkg {
			continue
		}
		items = append(items, ListItem{
			Package: pkg,
			Key:     fields[1],
			Type:    fields[2],
			Value:   strings.Join(fields[3:], " "),
		})
	}

	// Suppress unused import warning
	_ = os.DevNull

	return ListResult{
		Status:  "success",
		Package: pkg,
		Items:   items,
		Count:   len(items),
	}
}
