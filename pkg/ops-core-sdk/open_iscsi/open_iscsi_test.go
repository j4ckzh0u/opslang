package open_iscsi

import "testing"

func TestDiscoverEmpty(t *testing.T) {
	r := Discover("", 0)
	if r.Status != "failed" {
		t.Error("expected failure for empty portal")
	}
}

func TestLoginEmpty(t *testing.T) {
	r := Login("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty target")
	}
	r2 := Login("iqn.test", "")
	if r2.Status != "failed" {
		t.Error("expected failure for empty portal")
	}
}

func TestLogoutEmpty(t *testing.T) {
	r := Logout("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty target")
	}
}

func TestSetStartupMissing(t *testing.T) {
	r := SetStartup("", "", "")
	if r.Status != "failed" {
		t.Error("expected failure for missing args")
	}
}
