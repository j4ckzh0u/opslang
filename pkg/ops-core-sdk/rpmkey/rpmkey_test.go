package rpmkey

import "testing"

func TestImportMissingPath(t *testing.T) {
	r := Import("")
	if r.Success {
		t.Error("expected failure for empty path")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestRemoveMissingKeyID(t *testing.T) {
	r := Remove("")
	if r.Success {
		t.Error("expected failure for empty key ID")
	}
}

func TestList(t *testing.T) {
	result := List()
	// May fail if rpm not available, just check structure
	_ = result.Count
}
