package dpkg_selections

import (
	"testing"
)

func TestSetSelectionEmptyName(t *testing.T) {
	_, err := SetSelection("", "hold")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSetSelectionInvalidState(t *testing.T) {
	_, err := SetSelection("vim", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid state")
	}
}

func TestGetSelectionEmptyName(t *testing.T) {
	_, err := GetSelection("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSetSelectionNonDebian(t *testing.T) {
	_, err := SetSelection("vim", "hold")
	if err != nil {
		t.Skipf("dpkg not available: %v", err)
	}
}

func TestGetSelectionNonDebian(t *testing.T) {
	_, err := GetSelection("vim")
	if err != nil {
		t.Skipf("dpkg not available: %v", err)
	}
}

func TestResultStruct(t *testing.T) {
	r := Result{
		Status:   "success",
		Changed:  true,
		Package:  "vim",
		Previous: "install",
		Current:  "hold",
	}
	if r.Package != "vim" {
		t.Error("package mismatch")
	}
	if !r.Changed {
		t.Error("changed should be true")
	}
	if r.Previous != "install" {
		t.Error("previous mismatch")
	}
	if r.Current != "hold" {
		t.Error("current mismatch")
	}
}

func TestSelectionStruct(t *testing.T) {
	s := Selection{
		Package: "vim",
		State:   "hold",
	}
	if s.Package != "vim" {
		t.Error("package mismatch")
	}
	if s.State != "hold" {
		t.Error("state mismatch")
	}
}
