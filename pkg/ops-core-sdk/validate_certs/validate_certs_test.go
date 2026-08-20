package validate_certs

import "testing"

func TestValidateEmptyHost(t *testing.T) {
	r := Validate("", 0, 0)
	if r.Error == "" {
		t.Error("expected error for empty host")
	}
}

func TestValidateInvalidHost(t *testing.T) {
	r := Validate("nonexistent.invalid.host.example.com", 443, 0)
	if r.Error == "" {
		t.Error("expected error for invalid host")
	}
	if r.Valid {
		t.Error("expected not valid")
	}
}

func TestCheckExpiryInvalid(t *testing.T) {
	r := CheckExpiry("", 0, 30, 0)
	if r.Error == "" {
		t.Error("expected error for empty host")
	}
}
