// Package iptables provides Linux iptables firewall management operations.
package iptables

import (
	"fmt"
	"os/exec"
	"strings"
)

// Rule represents an iptables rule.
type Rule struct {
	Num     int    `json:"num"`
	Target  string `json:"target"`
	Proto   string `json:"proto"`
	Opt     string `json:"opt"`
	Source  string `json:"source"`
	Dest    string `json:"dest"`
	Extra   string `json:"extra"`
}

// ListResult represents the result of listing rules.
type ListResult struct {
	Chain string `json:"chain"`
	Rules []Rule `json:"rules"`
}

// ChainInfo represents an iptables chain.
type ChainInfo struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
	Packets string `json:"packets"`
	Bytes  string `json:"bytes"`
}

// ChainsResult represents the result of listing chains.
type ChainsResult struct {
	Chains []ChainInfo `json:"chains"`
}

// ActionResult represents the result of an iptables action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// List returns all rules in a chain.
func List(chain string) (*ListResult, error) {
	if chain == "" {
		chain = "INPUT"
	}

	out, err := exec.Command("iptables", "-L", chain, "--line-numbers", "-n").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables list failed: %w (output: %s)", err, string(out))
	}

	result := &ListResult{Chain: chain, Rules: make([]Rule, 0)}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Chain") || strings.HasPrefix(line, "target") || strings.HasPrefix(line, "num") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			rule := Rule{}
			fmt.Sscanf(fields[0], "%d", &rule.Num)
			rule.Target = fields[1]
			rule.Proto = fields[2]
			rule.Opt = fields[3]
			if len(fields) > 4 {
				rule.Source = fields[4]
			}
			if len(fields) > 5 {
				rule.Dest = fields[5]
			}
			if len(fields) > 6 {
				rule.Extra = strings.Join(fields[6:], " ")
			}
			result.Rules = append(result.Rules, rule)
		}
	}

	return result, nil
}

// Flush flushes all rules in a chain (or all chains if empty).
func Flush(chain string) (*ActionResult, error) {
	var cmd *exec.Cmd
	if chain == "" {
		cmd = exec.Command("iptables", "-F")
	} else {
		cmd = exec.Command("iptables", "-F", chain)
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("iptables flush failed: %w", err)
	}

	msg := "All chains flushed"
	if chain != "" {
		msg = fmt.Sprintf("Chain %s flushed", chain)
	}
	return &ActionResult{Changed: true, Message: msg}, nil
}

// AddRule adds a rule to a chain.
func AddRule(chain string, ruleSpec string) (*ActionResult, error) {
	if chain == "" {
		chain = "INPUT"
	}

	args := append([]string{"-A", chain}, strings.Fields(ruleSpec)...)
	cmd := exec.Command("iptables", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables add rule failed: %w (output: %s)", err, string(out))
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Added rule to %s: %s", chain, ruleSpec),
	}, nil
}

// DeleteRule deletes a rule by number from a chain.
func DeleteRule(chain string, num int) (*ActionResult, error) {
	if chain == "" {
		chain = "INPUT"
	}

	cmd := exec.Command("iptables", "-D", chain, fmt.Sprintf("%d", num))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables delete rule failed: %w (output: %s)", err, string(out))
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Deleted rule #%d from %s", num, chain),
	}, nil
}

// Save saves iptables rules (iptables-save).
func Save() (*ActionResult, error) {
	out, err := exec.Command("iptables-save").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables-save failed: %w", err)
	}

	return &ActionResult{
		Changed: false,
		Message: string(out),
	}, nil
}

// ListChains returns all chains with their policies.
func ListChains() (*ChainsResult, error) {
	out, err := exec.Command("iptables", "-L", "-n", "-v", "--line-numbers").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables list chains failed: %w", err)
	}

	result := &ChainsResult{Chains: make([]ChainInfo, 0)}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Chain ") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				chain := ChainInfo{
					Name:   fields[1],
					Policy: strings.Trim(fields[3], "()"),
				}
				if len(fields) >= 6 {
					chain.Packets = fields[4]
					chain.Bytes = fields[5]
				}
				result.Chains = append(result.Chains, chain)
			}
		}
	}

	return result, nil
}
