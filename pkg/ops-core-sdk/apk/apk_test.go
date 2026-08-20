package apk

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

func TestSearchEmptyName(t *testing.T) {
	_, err := Search("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestResultStruct(t *testing.T) {
	r := Result{
		Status:  "success",
		Changed: true,
		Package: "vim",
		Version: "9.0",
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
}

func TestInfoResultStruct(t *testing.T) {
	r := InfoResult{
		Status:      "success",
		Package:     "vim",
		Version:     "9.0",
		Description: "Vi IMproved",
		URL:         "https://www.vim.org",
		License:     "Vim",
	}
	if r.Package != "vim" {
		t.Error("package mismatch")
	}
	if r.Version != "9.0" {
		t.Error("version mismatch")
	}
}

func TestListNonAlpine(t *testing.T) {
	_, err := List()
	if err != nil {
		t.Skipf("apk not available: %v", err)
	}
}

func TestRepositoryNonAlpine(t *testing.T) {
	_, err := Repository()
	if err != nil {
		t.Skipf("not on Alpine: %v", err)
	}
}
