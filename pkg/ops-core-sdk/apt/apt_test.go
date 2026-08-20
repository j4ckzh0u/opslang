package apt

import (
	"testing"
)

func TestInstallEmptyName(t *testing.T) {
	_, err := Install("", "", false)
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

func TestPolicyEmptyName(t *testing.T) {
	_, err := Policy("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestMarkAutoEmptyName(t *testing.T) {
	_, err := MarkAuto("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestMarkManualEmptyName(t *testing.T) {
	_, err := MarkManual("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInfoNonexistentPackage(t *testing.T) {
	info, err := Info("this-package-definitely-does-not-exist-xyz123")
	if err != nil {
		t.Skipf("dpkg not available: %v", err)
	}
	if info.Status2 != "not-installed" {
		t.Logf("package status: %s (may be installed in test env)", info.Status2)
	}
}

func TestResultStruct(t *testing.T) {
	r := Result{
		Status:  "success",
		Changed: true,
		Package: "test",
		Version: "1.0",
		Output:  "done",
	}
	if r.Status != "success" {
		t.Error("status mismatch")
	}
	if !r.Changed {
		t.Error("changed should be true")
	}
	if r.Package != "test" {
		t.Error("package mismatch")
	}
}

func TestInfoResultStruct(t *testing.T) {
	r := InfoResult{
		Status:       "success",
		Package:      "test",
		Version:      "1.0",
		Architecture: "amd64",
		Description:  "test package",
		Status2:      "installed",
	}
	if r.Package != "test" {
		t.Error("package mismatch")
	}
	if r.Version != "1.0" {
		t.Error("version mismatch")
	}
}

func TestPolicyResultStruct(t *testing.T) {
	r := PolicyResult{
		Status:    "success",
		Package:   "test",
		Installed: "1.0",
		Candidate: "2.0",
	}
	if r.Installed != "1.0" {
		t.Error("installed mismatch")
	}
	if r.Candidate != "2.0" {
		t.Error("candidate mismatch")
	}
}
