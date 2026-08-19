package systemd

import (
	"testing"
)

func TestIsActive_RequiresUnit(t *testing.T) {
	_, err := IsActive("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestIsEnabled_RequiresUnit(t *testing.T) {
	_, err := IsEnabled("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestEnable_RequiresUnit(t *testing.T) {
	_, err := Enable("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestDisable_RequiresUnit(t *testing.T) {
	_, err := Disable("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestStart_RequiresUnit(t *testing.T) {
	_, err := Start("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestStop_RequiresUnit(t *testing.T) {
	_, err := Stop("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestRestart_RequiresUnit(t *testing.T) {
	_, err := Restart("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestReload_RequiresUnit(t *testing.T) {
	_, err := Reload("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestMask_RequiresUnit(t *testing.T) {
	_, err := Mask("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestUnmask_RequiresUnit(t *testing.T) {
	_, err := Unmask("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestShow_RequiresUnit(t *testing.T) {
	_, err := Show("")
	if err == nil {
		t.Error("expected error for empty unit name")
	}
}

func TestList_NoFilter(t *testing.T) {
	// List may fail if systemctl is not available, but should not panic
	result, err := List("")
	if err != nil {
		t.Skip("systemctl not available:", err)
	}
	if result.Count < 0 {
		t.Error("count should be >= 0")
	}
}

func TestList_WithType(t *testing.T) {
	result, err := List("service")
	if err != nil {
		t.Skip("systemctl not available:", err)
	}
	if result.Count < 0 {
		t.Error("count should be >= 0")
	}
}

func TestActionResultFields(t *testing.T) {
	r := ActionResult{
		Unit:    "test.service",
		Action:  "start",
		Changed: true,
		Message: "started",
	}
	if r.Unit != "test.service" {
		t.Error("unit mismatch")
	}
	if r.Action != "start" {
		t.Error("action mismatch")
	}
	if !r.Changed {
		t.Error("changed should be true")
	}
}

func TestStatusResultFields(t *testing.T) {
	r := StatusResult{
		Unit:      "test.service",
		Active:    true,
		Enabled:   true,
		Loaded:    true,
		State:     "active",
		LoadState: "loaded",
		SubState:  "running",
	}
	if !r.Active {
		t.Error("active should be true")
	}
	if !r.Enabled {
		t.Error("enabled should be true")
	}
}
