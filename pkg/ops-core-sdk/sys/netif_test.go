package sys

import (
	"strings"
	"testing"
)

// TestIsBusinessInterface pins the business-NIC rules against synthetic
// interfaces: no real machine needed.
func TestIsBusinessInterface(t *testing.T) {
	cases := []struct {
		desc  string
		name  string
		up    bool
		addrs int
		want  bool
	}{
		{"eth0 up with addr", "eth0", true, 1, true},
		{"ens192 up with addr", "ens192", true, 2, true},
		{"wlan0 up with addr", "wlan0", true, 1, true},
		{"docker bridge", "docker0", true, 1, false},
		{"docker-style bridge br-abc123", "br-a1b2c3d4e5f6", true, 1, false},
		{"short br-suffix stays visible", "br-x1", true, 1, true},
		{"real LAN bridge br-lan stays", "br-lan", true, 1, true},
		{"veth pair", "veth9f3a21", true, 1, false},
		{"kubernetes cni", "cni0", true, 1, false},
		{"calico pod iface", "cali1234ab", true, 1, false},
		{"flannel", "flannel.1", true, 1, false},
		{"libvirt bridge", "virbr0", true, 1, false},
		{"wireguard tunnel", "wg0", true, 1, false},
		{"openvpn tunnel", "tun0", true, 1, false},
		{"macos tunnel utun3", "utun3", true, 1, false},
		{"macos awdl", "awdl0", true, 1, false},
		{"macos llw", "llw0", true, 1, false},
		{"macos bridge100", "bridge100", true, 1, false},
		{"loopback linux", "lo", true, 1, false},
		{"loopback macos lo0", "lo0", true, 3, false},
		{"down interface", "eth1", false, 0, false},
		{"up but no addresses", "eth2", true, 0, false},
		{"empty name edge", "", true, 1, true}, // unknown names stay visible
	}
	for _, tc := range cases {
		if got := IsBusinessInterface(tc.name, tc.up, tc.addrs); got != tc.want {
			t.Errorf("%s: IsBusinessInterface(%q, %v, %d) = %v, want %v",
				tc.desc, tc.name, tc.up, tc.addrs, got, tc.want)
		}
	}
}

// TestGetNetInterfacesFiltered runs the real call and asserts invariants
// that must hold wherever it executes.
func TestGetNetInterfacesFiltered(t *testing.T) {
	ifaces, err := GetNetInterfaces()
	if err != nil {
		t.Fatalf("GetNetInterfaces() error = %v", err)
	}
	for _, iface := range ifaces {
		if isVirtualIFName(iface.Name) {
			t.Errorf("virtual interface leaked: %s", iface.Name)
		}
		if !iface.Up {
			t.Errorf("down interface leaked: %s", iface.Name)
		}
	}
	all, err := GetAllNetInterfaces()
	if err != nil {
		t.Fatalf("GetAllNetInterfaces() error = %v", err)
	}
	if len(all) < len(ifaces) {
		t.Errorf("raw list (%d) smaller than filtered (%d)", len(all), len(ifaces))
	}
	sawLo := false
	for _, iface := range all {
		// Linux names it "lo", macOS "lo0".
		if iface.Name == "lo" || strings.HasPrefix(iface.Name, "lo") {
			sawLo = true
		}
	}
	if !sawLo {
		t.Error("raw table must contain loopback; escape hatch is incomplete")
	}
}

// TestGetPrimaryIP checks the primary-IP selection contract on this
// machine: an IPv4 address is returned (global or, on locked-down
// machines, at least a fallback), never loopback when any global IP
// exists.
func TestGetPrimaryIP(t *testing.T) {
	result, err := GetPrimaryIP()
	if err != nil {
		t.Skipf("no primary IP on this machine: %v", err)
	}
	if result.Address == "" || result.Interface == "" {
		t.Fatalf("incomplete result: %+v", result)
	}
	if strings.HasPrefix(result.Address, "127.") {
		t.Errorf("primary IP must never be loopback: %+v", result)
	}
	if strings.Contains(result.Address, ":") {
		t.Errorf("primary IP must be IPv4: %+v", result)
	}
}
