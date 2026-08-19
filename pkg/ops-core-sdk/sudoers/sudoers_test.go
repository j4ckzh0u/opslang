package sudoers

import (
	"path/filepath"
	"testing"
)

func TestSetMissingArgs(t *testing.T) {
	r := Set("", "", "", false, "")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
	r2 := Set("test", "", "", false, "")
	if r2.Status != "failed" {
		t.Error("expected failure for empty user")
	}
}

func TestRemoveEmpty(t *testing.T) {
	r := Remove("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestInfoEmpty(t *testing.T) {
	r := Info("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestSetAndRemove(t *testing.T) {
	dir := t.TempDir()
	sudoersDir := filepath.Join(dir, "sudoers.d")

	r := Set("testuser", "deploy", "/usr/bin/systemctl restart nginx", true, sudoersDir)
	if r.Status != "success" {
		t.Errorf("expected success, got %s: %s", r.Status, r.Error)
	}

	info := Info("testuser", sudoersDir)
	if info.Status != "success" || !info.Exists {
		t.Error("expected file to exist")
	}

	// Remove
	rr := Remove("testuser", sudoersDir)
	if rr.Status != "success" || !rr.Changed {
		t.Errorf("expected changed, got status=%s changed=%v", rr.Status, rr.Changed)
	}
}
