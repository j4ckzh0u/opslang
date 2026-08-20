package dnf

import (
	"testing"
)

func TestInstallEmptyName(t *testing.T) {
	_, err := Install("", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemoveEmptyName(t *testing.T) {
	_, err := Remove("")
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

func TestSearchEmptyName(t *testing.T) {
	_, err := Search("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGroupInstallEmptyName(t *testing.T) {
	_, err := GroupInstall("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGroupRemoveEmptyName(t *testing.T) {
	_, err := GroupRemove("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestModuleEnableEmptySpec(t *testing.T) {
	_, err := ModuleEnable("")
	if err == nil {
		t.Fatal("expected error for empty spec")
	}
}

func TestResultStruct(t *testing.T) {
	r := Result{
		Status:  "success",
		Changed: true,
		Package: "vim",
		Version: "8.2",
		Output:  "installed",
	}
	if r.Status != "success" {
		t.Error("status mismatch")
	}
	if !r.Changed {
		t.Error("changed should be true")
	}
	if r.Package != "vim" {
		t.Error("package mismatch")
	}
	if r.Version != "8.2" {
		t.Error("version mismatch")
	}
}

func TestInfoResultStruct(t *testing.T) {
	r := InfoResult{
		Status:       "success",
		Package:      "vim",
		Version:      "8.2",
		Release:      "1.fc38",
		Architecture: "x86_64",
		Size:         "3.0 M",
		Summary:      "The VIM editor",
		License:      "Vim",
	}
	if r.Package != "vim" {
		t.Error("package mismatch")
	}
	if r.Version != "8.2" {
		t.Error("version mismatch")
	}
	if r.Architecture != "x86_64" {
		t.Error("arch mismatch")
	}
}

func TestRepoResultStruct(t *testing.T) {
	r := RepoResult{
		Status:  "success",
		ID:      "fedora",
		Name:    "Fedora 38",
		State:   "enabled",
		Enabled: true,
	}
	if r.ID != "fedora" {
		t.Error("id mismatch")
	}
	if !r.Enabled {
		t.Error("enabled should be true")
	}
}

func TestHistoryDefault(t *testing.T) {
	// Just verify the function doesn't panic with count=0
	// It will fail because dnf is not installed on test machine, but should not panic
	_, err := History(0)
	if err != nil {
		t.Skipf("dnf not available: %v", err)
	}
}
