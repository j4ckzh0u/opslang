package lvm_enhanced

import "testing"

func TestPVCreateEmpty(t *testing.T) {
	r := PVCreate("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestPVRemoveEmpty(t *testing.T) {
	r := PVRemove("", false)
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestVGCreateEmpty(t *testing.T) {
	r := VGCreate("", nil)
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
	r2 := VGCreate("vg1", nil)
	if r2.Status != "failed" {
		t.Error("expected failure for empty devices")
	}
}

func TestVGRemoveEmpty(t *testing.T) {
	r := VGRemove("", false)
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestVGExtendEmpty(t *testing.T) {
	r := VGExtend("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty args")
	}
}

func TestLVExtendEmpty(t *testing.T) {
	r := LVExtend("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty args")
	}
}

func TestLVExtendAllEmpty(t *testing.T) {
	r := LVExtendAll("")
	if r.Status != "failed" {
		t.Error("expected failure for empty path")
	}
}
