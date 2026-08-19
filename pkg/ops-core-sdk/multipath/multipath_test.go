package multipath

import "testing"

func TestAddMapEmpty(t *testing.T) {
	r := AddMap("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestRemoveMapEmpty(t *testing.T) {
	r := RemoveMap("")
	if r.Status != "failed" {
		t.Error("expected failure for empty device")
	}
}

func TestListPaths(t *testing.T) {
	r := ListPaths()
	_ = r.Status
}

func TestListMaps(t *testing.T) {
	r := ListMaps()
	_ = r.Status
}
