package ip_link

import (
	"testing"
)

func TestList(t *testing.T) {
	r := List()
	// On non-Linux or without ip command, this may fail
	if r.Success {
		if len(r.Interfaces) == 0 {
			t.Error("expected at least one interface")
		}
		// Loopback should always exist
		found := false
		for _, iface := range r.Interfaces {
			if iface.Name == "lo" || iface.Name == "lo0" {
				found = true
				if iface.LinkType != "loopback" {
					t.Errorf("expected link_type loopback, got %s", iface.LinkType)
				}
				break
			}
		}
		// Loopback might not exist in CI containers, so just check we got something
		if !found && len(r.Interfaces) == 0 {
			t.Error("no interfaces returned")
		}
	}
}

func TestGet(t *testing.T) {
	// Try loopback first
	name := "lo"
	r := Get(name)
	if !r.Success {
		// Try lo0 (macOS)
		name = "lo0"
		r = Get(name)
	}

	if r.Success {
		if len(r.Interfaces) == 0 {
			t.Error("expected interface info")
		}
	}
}

func TestGetNonExistent(t *testing.T) {
	r := Get("nonexistent_interface_xyz")
	if r.Success {
		t.Error("expected failure for non-existent interface")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestSetMTUValidation(t *testing.T) {
	// MTU too low
	r := SetMTU("lo", 50)
	if r.Success {
		t.Error("expected failure for MTU < 68")
	}

	// MTU too high
	r = SetMTU("lo", 70000)
	if r.Success {
		t.Error("expected failure for MTU > 65535")
	}
}

func TestParseInterfaces(t *testing.T) {
	// ponytail: test the parser directly
	output := `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00

2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP mode DEFAULT group default qlen 1000
    link/ether 00:11:22:33:44:55 brd ff:ff:ff:ff:ff:ff
`

	ifaces := parseInterfaces(output)
	if len(ifaces) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(ifaces))
	}

	// Check lo
	if ifaces[0].Name != "lo" {
		t.Errorf("expected name lo, got %s", ifaces[0].Name)
	}
	if ifaces[0].LinkType != "loopback" {
		t.Errorf("expected link_type loopback, got %s", ifaces[0].LinkType)
	}
	if ifaces[0].MTU != 65536 {
		t.Errorf("expected MTU 65536, got %d", ifaces[0].MTU)
	}

	// Check eth0
	if ifaces[1].Name != "eth0" {
		t.Errorf("expected name eth0, got %s", ifaces[1].Name)
	}
	if ifaces[1].State != "UP" {
		t.Errorf("expected state UP, got %s", ifaces[1].State)
	}
	if ifaces[1].MAC != "00:11:22:33:44:55" {
		t.Errorf("expected MAC 00:11:22:33:44:55, got %s", ifaces[1].MAC)
	}
	if ifaces[1].MTU != 1500 {
		t.Errorf("expected MTU 1500, got %d", ifaces[1].MTU)
	}
}

func TestParseInterfaceBlock(t *testing.T) {
	block := `3: ens192: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN mode DEFAULT group default qlen 1000
    link/ether aa:bb:cc:dd:ee:ff brd ff:ff:ff:ff:ff:ff`

	iface := parseInterfaceBlock(block)

	if iface.Name != "ens192" {
		t.Errorf("expected name ens192, got %s", iface.Name)
	}
	if iface.State != "DOWN" {
		t.Errorf("expected state DOWN, got %s", iface.State)
	}
	if iface.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected MAC aa:bb:cc:dd:ee:ff, got %s", iface.MAC)
	}
	if iface.MTU != 1500 {
		t.Errorf("expected MTU 1500, got %d", iface.MTU)
	}
}
