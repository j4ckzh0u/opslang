package lvg

import (
	"testing"
)

func TestActionResult(t *testing.T) {
	r := ActionResult{
		Name:    "vg0",
		Changed: true,
		Success: true,
	}
	if r.Name != "vg0" {
		t.Errorf("expected vg0, got %s", r.Name)
	}
	if !r.Changed {
		t.Error("expected changed")
	}
	if !r.Success {
		t.Error("expected success")
	}
}

func TestVGInfo(t *testing.T) {
	info := VGInfo{
		Name:    "vg0",
		PVCount: 2,
		LVCount: 3,
		VGSize:  "100G",
		VGFree:  "50G",
	}
	if info.Name != "vg0" {
		t.Errorf("expected vg0, got %s", info.Name)
	}
	if info.PVCount != 2 {
		t.Errorf("expected 2 PVs, got %d", info.PVCount)
	}
}

func TestCreate_EmptyName(t *testing.T) {
	_, err := Create("", []string{"/dev/sda1"})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestCreate_NoPVs(t *testing.T) {
	_, err := Create("vg0", []string{})
	if err == nil {
		t.Error("expected error for no physical volumes")
	}
}

func TestRemove_EmptyName(t *testing.T) {
	_, err := Remove("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestExtend_EmptyName(t *testing.T) {
	_, err := Extend("", []string{"/dev/sdb1"})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestExtend_NoPVs(t *testing.T) {
	_, err := Extend("vg0", []string{})
	if err == nil {
		t.Error("expected error for no physical volumes")
	}
}

func TestReduce_EmptyName(t *testing.T) {
	_, err := Reduce("", []string{"/dev/sda1"})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestReduce_NoPVs(t *testing.T) {
	_, err := Reduce("vg0", []string{})
	if err == nil {
		t.Error("expected error for no physical volumes")
	}
}

func TestActivate_EmptyName(t *testing.T) {
	_, err := Activate("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestDeactivate_EmptyName(t *testing.T) {
	_, err := Deactivate("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestGet_EmptyName(t *testing.T) {
	_, err := Get("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"42", 42},
		{"", 0},
		{"  5  ", 5},
	}

	for _, tt := range tests {
		result := parseInt(tt.input)
		if result != tt.expected {
			t.Errorf("parseInt(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}
