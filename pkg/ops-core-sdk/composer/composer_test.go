package composer

import "testing"

func TestInstallEmpty(t *testing.T) {
	r := Install("", false)
	if r.Status != "failed" {
		t.Error("expected failure for empty directory")
	}
}

func TestUpdateEmpty(t *testing.T) {
	r := Update("", false)
	if r.Status != "failed" {
		t.Error("expected failure for empty directory")
	}
}

func TestRequireMissing(t *testing.T) {
	r := Require("", "", "")
	if r.Status != "failed" {
		t.Error("expected failure for missing args")
	}
}

func TestRemoveMissing(t *testing.T) {
	r := Remove("", "")
	if r.Status != "failed" {
		t.Error("expected failure for missing args")
	}
}
