// Package nftables provides nftables firewall management operations.
// Equivalent to Ansible's nftables module. Replaces iptables.
package nftables

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result represents the result of an nft operation.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TableInfo represents a table.
type TableInfo struct {
	Family string `json:"family"`
	Name   string `json:"name"`
}

// ChainInfo represents a chain.
type ChainInfo struct {
	Family string `json:"family"`
	Table  string `json:"table"`
	Name   string `json:"name"`
	Type   string `json:"type,omitempty"`
	Hook   string `json:"hook,omitempty"`
	Prio   string `json:"priority,omitempty"`
}

// RuleInfo represents a rule.
type RuleInfo struct {
	Family  string `json:"family"`
	Table   string `json:"table"`
	Chain   string `json:"chain"`
	Handle  string `json:"handle"`
	Expr    string `json:"expression"`
}

// SetInfo represents a set.
type SetInfo struct {
	Family string `json:"family"`
	Table  string `json:"table"`
	Name   string `json:"name"`
	Type   string `json:"type,omitempty"`
	Flags  string `json:"flags,omitempty"`
}

func findNft() (string, error) {
	if p, err := exec.LookPath("nft"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("nft not found")
}

func runCmd(args ...string) (string, error) {
	nft, err := findNft()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(nft, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// AddTable creates a new table.
func AddTable(family string, name string) (Result, error) {
	if family == "" || name == "" {
		return Result{Status: "failed", Error: "family and name are required"}, fmt.Errorf("family and name are required")
	}
	out, err := runCmd("add", "table", family, name)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft add table: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// DeleteTable deletes a table.
func DeleteTable(family string, name string) (Result, error) {
	if family == "" || name == "" {
		return Result{Status: "failed", Error: "family and name are required"}, fmt.Errorf("family and name are required")
	}
	out, err := runCmd("delete", "table", family, name)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft delete table: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// ListTables lists all tables.
func ListTables() ([]TableInfo, error) {
	out, err := runCmd("list", "tables")
	if err != nil {
		return nil, fmt.Errorf("nft list tables: %w", err)
	}
	var results []TableInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: table ip filter
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			results = append(results, TableInfo{
				Family: parts[1],
				Name:   parts[2],
			})
		}
	}
	return results, nil
}

// AddChain creates a new chain.
func AddChain(family string, table string, name string, chainType string, hook string, priority string) (Result, error) {
	if family == "" || table == "" || name == "" {
		return Result{Status: "failed", Error: "family, table, and name are required"}, fmt.Errorf("family, table, and name are required")
	}
	args := []string{"add", "chain", family, table, name}
	if chainType != "" {
		args = append(args, "{", "type", chainType)
		if hook != "" {
			args = append(args, "hook", hook)
		}
		if priority != "" {
			args = append(args, "priority", priority)
		}
		args = append(args, ";", "}")
	}
	out, err := runCmd(args...)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft add chain: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// DeleteChain deletes a chain.
func DeleteChain(family string, table string, name string) (Result, error) {
	if family == "" || table == "" || name == "" {
		return Result{Status: "failed", Error: "family, table, and name are required"}, fmt.Errorf("family, table, and name are required")
	}
	out, err := runCmd("delete", "chain", family, table, name)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft delete chain: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// AddRule adds a rule to a chain.
func AddRule(family string, table string, chain string, expression string) (Result, error) {
	if family == "" || table == "" || chain == "" || expression == "" {
		return Result{Status: "failed", Error: "family, table, chain, and expression are required"}, fmt.Errorf("family, table, chain, and expression are required")
	}
	args := []string{"add", "rule", family, table, chain}
	// Expression is passed as raw args
	for _, part := range strings.Fields(expression) {
		args = append(args, part)
	}
	out, err := runCmd(args...)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft add rule: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// DeleteRule deletes a rule by handle.
func DeleteRule(family string, table string, chain string, handle string) (Result, error) {
	if family == "" || table == "" || chain == "" || handle == "" {
		return Result{Status: "failed", Error: "family, table, chain, and handle are required"}, fmt.Errorf("family, table, chain, and handle are required")
	}
	out, err := runCmd("delete", "rule", family, table, chain, "handle", handle)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft delete rule: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// FlushChain flushes all rules from a chain.
func FlushChain(family string, table string, chain string) (Result, error) {
	if family == "" || table == "" || chain == "" {
		return Result{Status: "failed", Error: "family, table, and chain are required"}, fmt.Errorf("family, table, and chain are required")
	}
	out, err := runCmd("flush", "chain", family, table, chain)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft flush chain: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// FlushTable flushes all chains from a table.
func FlushTable(family string, table string) (Result, error) {
	if family == "" || table == "" {
		return Result{Status: "failed", Error: "family and table are required"}, fmt.Errorf("family and table are required")
	}
	out, err := runCmd("flush", "table", family, table)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft flush table: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// FlushRuleset flushes the entire ruleset.
func FlushRuleset() (Result, error) {
	out, err := runCmd("flush", "ruleset")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft flush ruleset: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// ListRuleset returns the complete ruleset.
func ListRuleset() (string, error) {
	out, err := runCmd("list", "ruleset")
	if err != nil {
		return "", fmt.Errorf("nft list ruleset: %w", err)
	}
	return out, nil
}

// AddSet creates a new set.
func AddSet(family string, table string, name string, setType string, flags string) (Result, error) {
	if family == "" || table == "" || name == "" || setType == "" {
		return Result{Status: "failed", Error: "family, table, name, and type are required"}, fmt.Errorf("family, table, name, and type are required")
	}
	args := []string{"add", "set", family, table, name, "{", "type", setType}
	if flags != "" {
		args = append(args, "flags", flags)
	}
	args = append(args, ";", "}")
	out, err := runCmd(args...)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft add set: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// DeleteSet deletes a set.
func DeleteSet(family string, table string, name string) (Result, error) {
	if family == "" || table == "" || name == "" {
		return Result{Status: "failed", Error: "family, table, and name are required"}, fmt.Errorf("family, table, and name are required")
	}
	out, err := runCmd("delete", "set", family, table, name)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft delete set: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// AddElement adds an element to a set.
func AddElement(family string, table string, set string, element string) (Result, error) {
	if family == "" || table == "" || set == "" || element == "" {
		return Result{Status: "failed", Error: "family, table, set, and element are required"}, fmt.Errorf("family, table, set, and element are required")
	}
	out, err := runCmd("add", "element", family, table, set, "{", element, "}")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft add element: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// DeleteElement deletes an element from a set.
func DeleteElement(family string, table string, set string, element string) (Result, error) {
	if family == "" || table == "" || set == "" || element == "" {
		return Result{Status: "failed", Error: "family, table, set, and element are required"}, fmt.Errorf("family, table, set, and element are required")
	}
	out, err := runCmd("delete", "element", family, table, set, "{", element, "}")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("nft delete element: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Export exports ruleset in a given format (json, xml).
func Export(format string) (string, error) {
	if format == "" {
		format = "json"
	}
	out, err := runCmd("export", format)
	if err != nil {
		return "", fmt.Errorf("nft export: %w", err)
	}
	return out, nil
}
