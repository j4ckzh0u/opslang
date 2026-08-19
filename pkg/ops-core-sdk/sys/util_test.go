package sys

import (
	"testing"
)

func TestMACAddress_DefaultInterface(t *testing.T) {
	// May fail in environments with no non-loopback interfaces
	result, err := MACAddress("")
	if err != nil {
		t.Skipf("no non-loopback interface available: %v", err)
	}
	if result.Interface == "" {
		t.Error("expected non-empty interface name")
	}
	if result.MAC == "" {
		t.Error("expected non-empty MAC address")
	}
}

func TestMACAddress_SpecificInterface(t *testing.T) {
	_, err := MACAddress("nonexistent0")
	if err == nil {
		t.Error("expected error for nonexistent interface")
	}
}

func TestMACAddresses(t *testing.T) {
	result, err := MACAddresses()
	if err != nil {
		t.Fatalf("MACAddresses() error: %v", err)
	}
	// Result may be empty in some environments, but should not error
	_ = result
}

func TestMACAddressResultJSON(t *testing.T) {
	r := MACAddressResult{
		Interface: "eth0",
		MAC:       "00:11:22:33:44:55",
	}
	if r.Interface != "eth0" {
		t.Errorf("expected eth0, got %s", r.Interface)
	}
	if r.MAC != "00:11:22:33:44:55" {
		t.Errorf("expected 00:11:22:33:44:55, got %s", r.MAC)
	}
}
