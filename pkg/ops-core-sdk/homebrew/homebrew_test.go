package homebrew

import (
	"testing"
)

func TestInstallEmptyName(t *testing.T) {
	_, err := Install("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemoveEmptyName(t *testing.T) {
	_, err := Remove("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInfoEmptyName(t *testing.T) {
	_, err := Info("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestTapEmptyName(t *testing.T) {
	_, err := Tap("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUntapEmptyName(t *testing.T) {
	_, err := Untap("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestResultStruct(t *testing.T) {
	r := Result{
		Status:  "success",
		Changed: true,
		Package: "wget",
		Version: "1.21",
		Output:  "installed",
	}
	if r.Status != "success" {
		t.Error("status mismatch")
	}
	if !r.Changed {
		t.Error("changed should be true")
	}
}

func TestInfoResultStruct(t *testing.T) {
	r := InfoResult{
		Status:      "success",
		Package:     "wget",
		Version:     "1.21",
		Description: "Internet file retriever",
		Homepage:    "https://www.gnu.org/software/wget/",
		Installed:   true,
		Outdated:    false,
	}
	if r.Package != "wget" {
		t.Error("package mismatch")
	}
	if !r.Installed {
		t.Error("installed should be true")
	}
}

func TestNonMacOS(t *testing.T) {
	_, err := Install("wget", false)
	if err != nil {
		t.Skipf("brew not available: %v", err)
	}
}
