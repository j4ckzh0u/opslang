package flatpak

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

func TestRemoveEmptyName(t *testing.T) {
	r, err := Remove("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestUpdateEmptyName(t *testing.T) {
	r, err := Update("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestInfoEmptyName(t *testing.T) {
	r, err := Info("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestRunEmptyName(t *testing.T) {
	r, err := Run("", nil, false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestList(t *testing.T) {
	r, _ := List(false)
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestRepair(t *testing.T) {
	// Will fail if flatpak not installed, but should not panic
	r, _ := Repair(false)
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestAddRemoteEmpty(t *testing.T) {
	r, err := AddRemote("", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}
