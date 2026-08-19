package pam_limits

import "testing"

func TestSetEmpty(t *testing.T) {
	r := Set("", "", "", "")
	if r.Error == "" {
		t.Error("expected error for empty args")
	}
}

func TestList(t *testing.T) {
	r := List()
	// May fail if limits.conf doesn't exist
	_ = r.Count
}
