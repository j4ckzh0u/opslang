package mdadm

import "testing"

func TestCreateMissingDevice(t *testing.T) {
	r := Create("", "1", []string{"/dev/sda1"})
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestCreateMissingDevices(t *testing.T) {
	r := Create("/dev/md0", "1", nil)
	if r.Status != "failed" {
		t.Error("expected failure for empty devices")
	}
}

func TestDetailEmpty(t *testing.T) {
	r := Detail("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestDestroyEmpty(t *testing.T) {
	r := Destroy("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestScan(t *testing.T) {
	r := Scan()
	_ = r.Status // may be empty array on systems without RAID
}

func TestAddMissingArgs(t *testing.T) {
	r := Add("", "")
	if r.Status != "failed" {
		t.Error("expected failure for missing args")
	}
}

func TestRemoveMissingArgs(t *testing.T) {
	r := Remove("", "")
	if r.Status != "failed" {
		t.Error("expected failure for missing args")
	}
}
