package snap

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
	r, err := Remove("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestRefreshEmptyName(t *testing.T) {
	r, err := Refresh("", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestEnableEmptyName(t *testing.T) {
	r, err := Enable("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestDisableEmptyName(t *testing.T) {
	r, err := Disable("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestGetEmptyName(t *testing.T) {
	r, err := Get("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestSwitchEmpty(t *testing.T) {
	r, _ := Switch("", "stable")
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
	r2, _ := Switch("core", "")
	if r2.Status != "failed" {
		t.Errorf("expected status=failed for empty channel, got %s", r2.Status)
	}
}

func TestList(t *testing.T) {
	r, _ := List()
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestChanges(t *testing.T) {
	r, _ := Changes()
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}
