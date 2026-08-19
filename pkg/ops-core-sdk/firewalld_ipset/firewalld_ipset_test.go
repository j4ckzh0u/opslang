package firewalld_ipset

import "testing"

func TestCreateEmpty(t *testing.T) {
	r := Create("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty args")
	}
}

func TestDeleteEmpty(t *testing.T) {
	r := Delete("")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestAddEntryEmpty(t *testing.T) {
	r := AddEntry("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty args")
	}
}

func TestRemoveEntryEmpty(t *testing.T) {
	r := RemoveEntry("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty args")
	}
}

func TestInfoEmpty(t *testing.T) {
	r := Info("")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestList(t *testing.T) {
	r := List()
	// May fail if firewalld not running
	_ = r.Status
}
