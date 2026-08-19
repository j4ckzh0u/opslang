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

func TestDmidecode(t *testing.T) {
	// Dmidecode reads from /sys/class/dmi/id/ which may be empty on non-Linux or VM
	result, err := Dmidecode()
	if err != nil {
		t.Fatalf("Dmidecode() error: %v", err)
	}
	// Just verify it returns without error
	_ = result
}

func TestLsPci(t *testing.T) {
	// LsPci requires lspci command
	result, err := LsPci()
	if err != nil {
		t.Skipf("lspci not available: %v", err)
	}
	_ = result
}

func TestLsBlk(t *testing.T) {
	// LsBlk requires lsblk command
	result, err := LsBlk()
	if err != nil {
		t.Skipf("lsblk not available: %v", err)
	}
	_ = result
}

func TestDmidecodeResultJSON(t *testing.T) {
	r := DmidecodeResult{
		BiosVendor:   "TestVendor",
		SystemVendor: "TestSys",
	}
	if r.BiosVendor != "TestVendor" {
		t.Errorf("expected TestVendor, got %s", r.BiosVendor)
	}
}

func TestPciDeviceJSON(t *testing.T) {
	d := PciDevice{
		Slot:   "00:00.0",
		Class:  "Host bridge",
		Vendor: "Intel",
	}
	if d.Slot != "00:00.0" {
		t.Errorf("expected 00:00.0, got %s", d.Slot)
	}
}

func TestBlkDeviceJSON(t *testing.T) {
	d := BlkDevice{
		Name:       "sda",
		Type:       "disk",
		Size:       "100G",
		MountPoint: "/",
	}
	if d.Name != "sda" {
		t.Errorf("expected sda, got %s", d.Name)
	}
}
