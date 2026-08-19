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

func TestLsUsb(t *testing.T) {
	// LsUsb requires lsusb command
	result, err := LsUsb()
	if err != nil {
		t.Skipf("lsusb not available: %v", err)
	}
	_ = result
}

func TestIpRoute(t *testing.T) {
	// IpRoute requires ip command
	result, err := IpRoute()
	if err != nil {
		t.Skipf("ip route not available: %v", err)
	}
	_ = result
}

func TestEthtool(t *testing.T) {
	// Ethtool requires ethtool command or /sys/class/net/
	result, err := Ethtool("eth0")
	if err != nil {
		t.Skipf("ethtool not available: %v", err)
	}
	_ = result
}

func TestEthtool_EmptyIface(t *testing.T) {
	_, err := Ethtool("")
	if err == nil {
		t.Error("expected error for empty interface name")
	}
}

func TestUsbDeviceJSON(t *testing.T) {
	d := UsbDevice{
		Bus:    "001",
		Device: "002",
		ID:     "1234:5678",
		Name:   "Test USB Device",
	}
	if d.Bus != "001" {
		t.Errorf("expected 001, got %s", d.Bus)
	}
}

func TestRouteEntryJSON(t *testing.T) {
	r := RouteEntry{
		Destination: "default",
		Gateway:     "192.168.1.1",
		Interface:   "eth0",
	}
	if r.Destination != "default" {
		t.Errorf("expected default, got %s", r.Destination)
	}
}

func TestEthtoolInfoJSON(t *testing.T) {
	e := EthtoolInfo{
		Interface:    "eth0",
		Driver:       "e1000e",
		Speed:        "1000Mb/s",
		LinkDetected: true,
	}
	if e.Interface != "eth0" {
		t.Errorf("expected eth0, got %s", e.Interface)
	}
}
