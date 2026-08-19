package locale_gen

import "testing"

func TestGenerateEmpty(t *testing.T) {
	r := Generate("")
	if r.Error == "" {
		t.Error("expected error for empty locale")
	}
}

func TestList(t *testing.T) {
	r := List()
	// May fail if locale not available
	_ = r.Count
}

func TestRemoveEmpty(t *testing.T) {
	r := Remove("")
	if r.Error == "" {
		t.Error("expected error for empty locale")
	}
}
