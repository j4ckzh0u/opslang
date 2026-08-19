package swap

import "testing"

func TestInfo(t *testing.T) {
	r := Info()
	// Should succeed on Linux
	if r.Status == "success" {
		if r.Total == 0 && len(r.Entries) > 0 {
			t.Error("entries present but total is 0")
		}
	}
}

func TestCreateInvalidArgs(t *testing.T) {
	r := Create("", 0)
	if r.Status != "failed" {
		t.Error("expected failure for empty args")
	}
}

func TestCreateNegativeSize(t *testing.T) {
	r := Create("/tmp/testswap", -1)
	if r.Status != "failed" {
		t.Error("expected failure for negative size")
	}
}

func TestEnableEmpty(t *testing.T) {
	r := Enable("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestDisableEmpty(t *testing.T) {
	r := Disable("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}
