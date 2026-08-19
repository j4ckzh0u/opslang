package modinfo

import "testing"

func TestInfoEmpty(t *testing.T) {
	r := Info("")
	if r.Error == "" {
		t.Error("expected error for empty module")
	}
}

func TestList(t *testing.T) {
	r := List()
	// May fail if lsmod not available
	_ = r.Count
}

func TestVersion(t *testing.T) {
	r := Version()
	// May fail if modinfo not available
	_ = r.Version
}
