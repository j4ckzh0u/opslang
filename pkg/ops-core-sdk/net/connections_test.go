package opsnet

import (
	"strings"
	"testing"
)

// The socket table is read from the live kernel — every assertion runs
// against real sockets on the test host.
func TestConnectionsLiveTable(t *testing.T) {
	conns, err := Connections("inet")
	if err != nil {
		t.Fatalf("Connections(inet): %v", err)
	}
	if len(conns) == 0 {
		t.Fatal("a live host always has loopback sockets; got an empty table")
	}

	// Every entry must carry a well-formed local address and protocol.
	for _, c := range conns {
		if c.Protocol != "tcp" && c.Protocol != "udp" {
			t.Errorf("protocol must be tcp|udp, got %q", c.Protocol)
		}
		if !strings.Contains(c.LocalAddr, ":") {
			t.Errorf("local address must be ip:port, got %q", c.LocalAddr)
		}
		if c.Pid > 0 && c.ProcessName == "" {
			// Name resolution is best-effort (a process can exit between
			// the two reads); on a quiet test host this is rare but not
			// impossible, so only log it.
			t.Logf("pid %d has no resolved name (may have exited)", c.Pid)
		}
	}

	// On macOS/CI the runner listens on loopback; at least one entry must
	// be attributable to this very test binary's process family. On some
	// CI sandboxes pid attribution is restricted, so this is a soft check:
	// we verify that at least one connection has a status string, which
	// proves the table was really parsed.
	statuses := 0
	for _, c := range conns {
		if c.Status != "" {
			statuses++
		}
	}
	if statuses == 0 {
		t.Error("no connection carried a status; socket table parse looks broken")
	}
}

func TestConnectionsTCPOnlyHasNoUDP(t *testing.T) {
	conns, err := Connections("tcp")
	if err != nil {
		t.Fatalf("Connections(tcp): %v", err)
	}
	for _, c := range conns {
		if c.Protocol != "tcp" {
			t.Errorf("kind=tcp must not return %s sockets", c.Protocol)
		}
	}
}

func TestConnectionsRejectsInvalidKind(t *testing.T) {
	for _, kind := range []string{"sctp", "raw", "inet9", "unix", "all"} {
		if _, err := Connections(kind); err == nil {
			t.Errorf("Connections(%q) must be rejected", kind)
		}
	}
	// Whitespace is tolerated (trimmed), empty kind defaults to inet.
	if _, err := Connections(" tcp "); err != nil {
		t.Errorf("Connections(\" tcp \") should trim to tcp: %v", err)
	}
	if _, err := Connections(""); err != nil {
		t.Errorf("Connections(\"\") should default to inet: %v", err)
	}
}

func TestFormatAddrIPv6Bracketed(t *testing.T) {
	if got := formatAddr("::1", 8080); got != "[::1]:8080" {
		t.Errorf("IPv6 must be bracketed, got %q", got)
	}
	if got := formatAddr("10.0.0.1", 22); got != "10.0.0.1:22" {
		t.Errorf("IPv4 must stay plain, got %q", got)
	}
}
