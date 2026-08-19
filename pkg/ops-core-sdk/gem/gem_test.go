package gem

import (
	"testing"
)

func TestInstallEmptyName(t *testing.T) {
	r, err := Install("", "", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestUninstallEmptyName(t *testing.T) {
	r, err := Uninstall("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestUpdateEmptyName(t *testing.T) {
	r, err := Update("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestInfoEmptyName(t *testing.T) {
	r, err := Info("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestList(t *testing.T) {
	r, _ := List()
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestVersion(t *testing.T) {
	r, _ := Version()
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}
