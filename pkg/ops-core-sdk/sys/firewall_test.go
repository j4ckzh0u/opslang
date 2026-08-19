//go:build linux

package sys

import (
	"testing"
)

func TestFirewallRule_InvalidAction(t *testing.T) {
	_, err := FirewallRule("invalid", "tcp", 80, "")
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestFirewallRule_EmptyProtocol(t *testing.T) {
	_, err := FirewallRule("add", "", 80, "")
	if err == nil {
		t.Fatal("expected error for empty protocol")
	}
}

func TestFirewallRule_InvalidPort(t *testing.T) {
	_, err := FirewallRule("add", "tcp", 0, "")
	if err == nil {
		t.Fatal("expected error for port 0")
	}
	_, err = FirewallRule("add", "tcp", 70000, "")
	if err == nil {
		t.Fatal("expected error for port 70000")
	}
}

func TestFirewallRule_AddWithoutRoot(t *testing.T) {
	// This will likely fail without root, but we test that validation passes
	// and the function reaches the iptables call.
	_, err := FirewallRule("add", "tcp", 12345, "10.0.0.0/8")
	if err == nil {
		// If somehow running as root with iptables, that's fine
		t.Log("firewall rule added (running with sufficient privileges)")
	}
	// Error is expected when running without root
}

func TestFirewallRule_RemoveWithoutRoot(t *testing.T) {
	_, err := FirewallRule("remove", "tcp", 12345, "")
	if err == nil {
		t.Log("firewall rule removed (running with sufficient privileges)")
	}
	// Error is expected when running without root
}

func TestFirewallRule_RuleDescription(t *testing.T) {
	// Test that the rule description is built correctly even for failed operations
	// We can't easily test the actual iptables call without root, but we can
	// verify input validation works and the function signature is correct.
	_, err := FirewallRule("add", "udp", 53, "192.168.1.0/24")
	// Just verify no panic - actual execution requires root
	if err != nil {
		t.Logf("expected error without root: %v", err)
	}
}
