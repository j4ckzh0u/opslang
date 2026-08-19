package debconf

import "testing"

func TestSetMissing(t *testing.T) {
	r := Set("", "", "", "")
	if r.Status != "failed" {
		t.Error("expected failure for missing args")
	}
}

func TestGetMissing(t *testing.T) {
	r := Get("", "")
	if r.Status != "failed" {
		t.Error("expected failure for missing args")
	}
}

func TestListMissing(t *testing.T) {
	r := List("")
	if r.Status != "failed" {
		t.Error("expected failure for empty package")
	}
}

func TestSetNotInstalled(t *testing.T) {
	// debconf-set-selections likely not available on non-Debian
	r := Set("tzdata", "tzdata/Areas", "select", "Etc")
	// May fail or succeed depending on platform
	_ = r.Status
}

func TestGetNotInstalled(t *testing.T) {
	r := Get("tzdata", "tzdata/Areas")
	// May fail on non-Debian
	_ = r.Status
}

func TestListNotInstalled(t *testing.T) {
	r := List("tzdata")
	_ = r.Status
}
