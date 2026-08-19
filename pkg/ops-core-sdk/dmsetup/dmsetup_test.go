package dmsetup

import "testing"

func TestCreateMissingArgs(t *testing.T) {
	r := Create("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
	r2 := Create("test", "")
	if r2.Status != "failed" {
		t.Error("expected failure for empty table")
	}
}

func TestRemoveEmpty(t *testing.T) {
	r := Remove("")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestInfoEmpty(t *testing.T) {
	r := Info("")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestSuspendEmpty(t *testing.T) {
	r := Suspend("")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestResumeEmpty(t *testing.T) {
	r := Resume("")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestList(t *testing.T) {
	r := List()
	_ = r.Status // may have no devices
}
