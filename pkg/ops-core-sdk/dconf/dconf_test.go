package dconf

import "testing"

func TestReadEmpty(t *testing.T) {
	r := Read("")
	if r.Error == "" {
		t.Error("expected error for empty key")
	}
}

func TestWriteEmpty(t *testing.T) {
	r := Write("", "")
	if r.Error == "" {
		t.Error("expected error for empty key")
	}
}

func TestWriteEmptyValue(t *testing.T) {
	r := Write("test", "")
	if r.Error == "" {
		t.Error("expected error for empty value")
	}
}

func TestList(t *testing.T) {
	r := List("")
	// May fail if dconf not available
	_ = r.Count
}

func TestResetEmpty(t *testing.T) {
	r := Reset("")
	if r.Error == "" {
		t.Error("expected error for empty key")
	}
}
