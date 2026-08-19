package aptkey

import "testing"

func TestAddMissingURL(t *testing.T) {
	r := Add("", "")
	if r.Success {
		t.Error("expected failure for empty URL")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestAddFromKeyMissingPath(t *testing.T) {
	r := AddFromKey("", "")
	if r.Success {
		t.Error("expected failure for empty path")
	}
}

func TestRemoveMissingKeyID(t *testing.T) {
	r := Remove("", "")
	if r.Success {
		t.Error("expected failure for empty key ID")
	}
}

func TestList(t *testing.T) {
	result := List()
	// May fail if apt-key not available
	_ = result.Count
}
