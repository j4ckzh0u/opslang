// Package pamd provides PAM (Pluggable Authentication Modules) configuration management.
// All operations parse and modify /etc/pam.d/ service files directly.
package pamd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const pamDir = "/etc/pam.d"

// Rule represents a single PAM rule line.
type Rule struct {
	Type     string `json:"type"`     // account, auth, password, session
	Control  string `json:"control"`  // required, sufficient, etc.
	Module   string `json:"module"`   // module path
	Args     string `json:"args"`     // module arguments
	LineNum  int    `json:"line_num"` // line number in file
	RawLine  string `json:"raw_line"` // original line
}

// ServiceResult is returned by Get.
type ServiceResult struct {
	Service string `json:"service"`
	Path    string `json:"path"`
	Exists  bool   `json:"exists"`
	Rules   []Rule `json:"rules"`
}

// ListResult is returned by List.
type ListResult struct {
	Services []string `json:"services"`
	Count    int      `json:"count"`
}

// ActionResult is returned by mutating operations.
type ActionResult struct {
	Service  string `json:"service"`
	Success  bool   `json:"success"`
	Changed  bool   `json:"changed"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// ValidateResult is returned by Validate.
type ValidateResult struct {
	Service string `json:"service"`
	Valid   bool   `json:"valid"`
	Errors  []string `json:"errors"`
}

// BackupResult is returned by Backup.
type BackupResult struct {
	Service    string `json:"service"`
	BackupPath string `json:"backup_path"`
	Success    bool   `json:"success"`
}

func parseRules(content string) []Rule {
	var rules []Rule
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) < 3 {
			continue
		}
		r := Rule{
			Type:    fields[0],
			Control: fields[1],
			Module:  fields[2],
			LineNum: i + 1,
			RawLine: line,
		}
		if len(fields) > 3 {
			r.Args = strings.Join(fields[3:], " ")
		}
		rules = append(rules, r)
	}
	return rules
}

func servicePath(service string) string {
	return filepath.Join(pamDir, service)
}

// Get returns all PAM rules for a service.
func Get(service string) (ServiceResult, error) {
	if service == "" {
		return ServiceResult{}, fmt.Errorf("service name is required")
	}
	p := servicePath(service)
	info, err := os.Stat(p)
	if err != nil || info.IsDir() {
		return ServiceResult{Service: service, Path: p, Exists: false}, nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return ServiceResult{Service: service, Path: p, Exists: true}, err
	}
	rules := parseRules(string(data))
	return ServiceResult{Service: service, Path: p, Exists: true, Rules: rules}, nil
}

// List returns all PAM service names.
func List() (ListResult, error) {
	entries, err := os.ReadDir(pamDir)
	if err != nil {
		return ListResult{}, fmt.Errorf("failed to read %s: %w", pamDir, err)
	}
	var services []string
	for _, e := range entries {
		if !e.IsDir() {
			services = append(services, e.Name())
		}
	}
	return ListResult{Services: services, Count: len(services)}, nil
}

// AddRule appends a new rule to a PAM service file.
func AddRule(service, ruleType, control, module, args string) (ActionResult, error) {
	start := time.Now()
	res := ActionResult{Service: service}
	if service == "" || ruleType == "" || control == "" || module == "" {
		return res, fmt.Errorf("service, type, control, and module are required")
	}
	p := servicePath(service)
	data, err := os.ReadFile(p)
	if err != nil {
		return res, fmt.Errorf("service %s not found: %w", service, err)
	}
	newLine := ruleType + "\t" + control + "\t" + module
	if args != "" {
		newLine += "\t" + args
	}
	content := strings.TrimRight(string(data), "\n") + "\n" + newLine + "\n"
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		return res, err
	}
	res.Success = true
	res.Changed = true
	res.Duration = time.Since(start).Milliseconds()
	return res, nil
}

// RemoveRule removes rules matching the given type and module.
func RemoveRule(service, ruleType, module string) (ActionResult, error) {
	start := time.Now()
	res := ActionResult{Service: service}
	if service == "" || ruleType == "" || module == "" {
		return res, fmt.Errorf("service, type, and module are required")
	}
	p := servicePath(service)
	data, err := os.ReadFile(p)
	if err != nil {
		return res, fmt.Errorf("service %s not found: %w", service, err)
	}
	lines := strings.Split(string(data), "\n")
	var kept []string
	changed := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			kept = append(kept, line)
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 3 && fields[0] == ruleType && fields[2] == module {
			changed = true
			continue
		}
		kept = append(kept, line)
	}
	if !changed {
		res.Success = true
		res.Changed = false
		res.Duration = time.Since(start).Milliseconds()
		return res, nil
	}
	if err := os.WriteFile(p, []byte(strings.Join(kept, "\n")), 0644); err != nil {
		return res, err
	}
	res.Success = true
	res.Changed = true
	res.Duration = time.Since(start).Milliseconds()
	return res, nil
}

// ModifyRule modifies the control/args of a rule matching type+module.
func ModifyRule(service, ruleType, module, newControl, newArgs string) (ActionResult, error) {
	start := time.Now()
	res := ActionResult{Service: service}
	if service == "" || ruleType == "" || module == "" {
		return res, fmt.Errorf("service, type, and module are required")
	}
	p := servicePath(service)
	data, err := os.ReadFile(p)
	if err != nil {
		return res, fmt.Errorf("service %s not found: %w", service, err)
	}
	lines := strings.Split(string(data), "\n")
	changed := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 3 && fields[0] == ruleType && fields[2] == module {
			newLine := ruleType + "\t" + newControl + "\t" + module
			if newArgs != "" {
				newLine += "\t" + newArgs
			}
			lines[i] = newLine
			changed = true
		}
	}
	if !changed {
		res.Success = true
		res.Changed = false
		res.Duration = time.Since(start).Milliseconds()
		return res, nil
	}
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return res, err
	}
	res.Success = true
	res.Changed = true
	res.Duration = time.Since(start).Milliseconds()
	return res, nil
}

// Validate checks a PAM service file for basic structural validity.
func Validate(service string) (ValidateResult, error) {
	if service == "" {
		return ValidateResult{}, fmt.Errorf("service name is required")
	}
	p := servicePath(service)
	data, err := os.ReadFile(p)
	if err != nil {
		return ValidateResult{Service: service, Valid: false, Errors: []string{"file not found"}}, nil
	}
	var errs []string
	validTypes := map[string]bool{"account": true, "auth": true, "password": true, "session": true, "-account": true, "-auth": true, "-password": true, "-session": true}
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) < 3 {
			errs = append(errs, fmt.Sprintf("line %d: too few fields", i+1))
			continue
		}
		if !validTypes[fields[0]] {
			errs = append(errs, fmt.Sprintf("line %d: unknown type %q", i+1, fields[0]))
		}
	}
	return ValidateResult{Service: service, Valid: len(errs) == 0, Errors: errs}, nil
}

// Backup creates a backup of a PAM service file.
func Backup(service, backupDir string) (BackupResult, error) {
	if service == "" {
		return BackupResult{}, fmt.Errorf("service name is required")
	}
	if backupDir == "" {
		backupDir = "/tmp"
	}
	p := servicePath(service)
	data, err := os.ReadFile(p)
	if err != nil {
		return BackupResult{}, fmt.Errorf("service %s not found: %w", service, err)
	}
	ts := time.Now().Format("20060102150405")
	bp := filepath.Join(backupDir, fmt.Sprintf("%s.pam.bak.%s", service, ts))
	if err := os.WriteFile(bp, data, 0644); err != nil {
		return BackupResult{}, err
	}
	return BackupResult{Service: service, BackupPath: bp, Success: true}, nil
}
