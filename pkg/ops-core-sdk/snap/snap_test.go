package snap

import (
	"testing"
)

func TestActionResult(t *testing.T) {
	r := ActionResult{
		Name:    "test-snap",
		Channel: "stable",
		Changed: true,
		Success: true,
	}
	if r.Name != "test-snap" {
		t.Errorf("expected test-snap, got %s", r.Name)
	}
	if !r.Changed {
		t.Error("expected changed")
	}
	if !r.Success {
		t.Error("expected success")
	}
}

func TestSnapInfo(t *testing.T) {
	info := SnapInfo{
		Name:     "test-snap",
		Version:  "1.0.0",
		Rev:      "123",
		Tracking: "stable",
		Publisher: "test",
	}
	if info.Name != "test-snap" {
		t.Errorf("expected test-snap, got %s", info.Name)
	}
}

func TestInstall_EmptyName(t *testing.T) {
	_, err := Install("", "stable", false)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRemove_EmptyName(t *testing.T) {
	_, err := Remove("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRefresh_EmptyName(t *testing.T) {
	_, err := Refresh("", "stable")
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

func TestEnable_EmptyName(t *testing.T) {
	_, err := Enable("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestDisable_EmptyName(t *testing.T) {
	_, err := Disable("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestSwitch_EmptyName(t *testing.T) {
	_, err := Switch("", "stable")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestSwitch_EmptyChannel(t *testing.T) {
	_, err := Switch("test-snap", "")
	if err == nil {
		t.Error("expected error for empty channel")
	}
}

func TestList(t *testing.T) {
	// May fail if snap is not installed
	result, err := List()
	if err != nil {
		t.Skipf("snap not available: %v", err)
	}
	_ = result
}

func TestChanges(t *testing.T) {
	// May fail if snap is not installed
	result, err := Changes()
	if err != nil {
		t.Skipf("snap not available: %v", err)
	}
	_ = result
}

func TestListResultJSON(t *testing.T) {
	r := ListResult{
		Snaps: []SnapInfo{
			{Name: "test", Version: "1.0"},
		},
	}
	json, err := r.JSON()
	if err != nil {
		t.Errorf("JSON() error: %v", err)
	}
	if json == "" {
		t.Error("expected non-empty JSON")
	}
}
