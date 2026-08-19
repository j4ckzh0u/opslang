package uri_ext

import "testing"

func TestPatchEmptyURL(t *testing.T) {
	r := Patch("", nil, nil, 5)
	if r.Status != "failed" {
		t.Error("expected failure for empty URL")
	}
}

func TestDeleteEmptyURL(t *testing.T) {
	r := Delete("", nil, 5)
	if r.Status != "failed" {
		t.Error("expected failure for empty URL")
	}
}

func TestHeadEmptyURL(t *testing.T) {
	r := Head("", nil, 5)
	if r.Status != "failed" {
		t.Error("expected failure for empty URL")
	}
}

func TestOptionsEmptyURL(t *testing.T) {
	r := Options("", nil, 5)
	if r.Status != "failed" {
		t.Error("expected failure for empty URL")
	}
}

func TestPatchInvalidURL(t *testing.T) {
	r := Patch("not-a-valid-url", []byte("{}"), nil, 2)
	if r.Status != "failed" {
		t.Error("expected failure for invalid URL")
	}
}

func TestDeleteInvalidURL(t *testing.T) {
	r := Delete("not-a-valid-url", nil, 2)
	if r.Status != "failed" {
		t.Error("expected failure for invalid URL")
	}
}
