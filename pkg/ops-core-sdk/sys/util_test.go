package sys

import (
	"strings"
	"testing"
)

func TestUUID(t *testing.T) {
	result, err := UUID()
	if err != nil {
		t.Fatal(err)
	}
	if result.UUID == "" {
		t.Error("UUID should not be empty")
	}
	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	parts := strings.Split(result.UUID, "-")
	if len(parts) != 5 {
		t.Errorf("UUID should have 5 parts, got %d: %s", len(parts), result.UUID)
	}
	// Version should be 4
	if len(parts[2]) > 0 && parts[2][0] != '4' {
		t.Errorf("UUID version should be 4, got %c in %s", parts[2][0], result.UUID)
	}
}

func TestUUID_Unique(t *testing.T) {
	r1, _ := UUID()
	r2, _ := UUID()
	if r1.UUID == r2.UUID {
		t.Error("two UUIDs should be unique")
	}
}

func TestRandomPassword_DefaultLength(t *testing.T) {
	result, err := RandomPassword(0, true, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Length < 8 {
		t.Errorf("password length should be at least 8, got %d", result.Length)
	}
	if result.Password == "" {
		t.Error("password should not be empty")
	}
}

func TestRandomPassword_SpecifiedLength(t *testing.T) {
	result, err := RandomPassword(20, true, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Length != 20 {
		t.Errorf("password length should be 20, got %d", result.Length)
	}
}

func TestRandomPassword_MinLength(t *testing.T) {
	result, err := RandomPassword(4, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	// Should be bumped to minimum 8
	if result.Length < 8 {
		t.Errorf("password length should be at least 8, got %d", result.Length)
	}
}

func TestRandomPassword_CharacterSets(t *testing.T) {
	// All character sets enabled
	result, err := RandomPassword(50, true, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Password == "" {
		t.Error("password should not be empty")
	}

	// Only lowercase
	result2, err := RandomPassword(20, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if result2.Password == "" {
		t.Error("password should not be empty")
	}
}

func TestRandomPassword_Unique(t *testing.T) {
	r1, _ := RandomPassword(20, true, true, true)
	r2, _ := RandomPassword(20, true, true, true)
	if r1.Password == r2.Password {
		t.Error("two passwords should be unique")
	}
}

func TestUUIDResultFields(t *testing.T) {
	r := UUIDResult{UUID: "test-uuid-123"}
	if r.UUID != "test-uuid-123" {
		t.Error("UUID mismatch")
	}
}

func TestPasswordResultFields(t *testing.T) {
	r := PasswordResult{Password: "test123", Length: 7}
	if r.Password != "test123" {
		t.Error("password mismatch")
	}
	if r.Length != 7 {
		t.Error("length mismatch")
	}
}
