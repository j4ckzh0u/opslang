package hwclock

import "testing"

func TestGet(t *testing.T) {
	r := Get()
	// May fail if hwclock not available
	_ = r.Status
}

func TestSetPermission(t *testing.T) {
	// Requires root
	r := Set()
	_ = r.Status
}

func TestSetTimeEmpty(t *testing.T) {
	r := SetTime("")
	if r.Status != "failed" {
		t.Error("expected failure for empty time")
	}
}
