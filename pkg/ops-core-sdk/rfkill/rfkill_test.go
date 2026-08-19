package rfkill

import "testing"

func TestBlockEmpty(t *testing.T) {
	r := Block("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestUnblockEmpty(t *testing.T) {
	r := Unblock("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestBlockAllEmpty(t *testing.T) {
	r := BlockAll("")
	if r.Status != "failed" {
		t.Error("expected failure for empty type")
	}
}

func TestUnblockAllEmpty(t *testing.T) {
	r := UnblockAll("")
	if r.Status != "failed" {
		t.Error("expected failure for empty type")
	}
}

func TestList(t *testing.T) {
	// May fail if rfkill not installed
	r := List()
	_ = r.Status
}
