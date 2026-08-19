package yarn

import "testing"

func TestInstallEmpty(t *testing.T) {
	r := Install("", "", false)
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestRemoveEmpty(t *testing.T) {
	r := Remove("", false)
	if r.Status != "failed" {
		t.Error("expected failure for empty name")
	}
}

func TestGlobalEmpty(t *testing.T) {
	r := Global("")
	if r.Status != "failed" {
		t.Error("expected failure for empty directory")
	}
}
