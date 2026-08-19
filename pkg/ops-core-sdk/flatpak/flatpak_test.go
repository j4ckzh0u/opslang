package flatpak

import (
	"encoding/json"
	"testing"
)

func TestInstall_EmptyName(t *testing.T) {
	_, err := Install("", "", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemove_EmptyName(t *testing.T) {
	_, err := Remove("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdate_EmptyName(t *testing.T) {
	_, err := Update("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInfo_EmptyName(t *testing.T) {
	_, err := Info("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRun_EmptyName(t *testing.T) {
	_, err := Run("", nil, false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Name: "org.gnome.Calculator", Changed: true, Success: true}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestFlatpakInfoJSON(t *testing.T) {
	info := FlatpakInfo{
		Name:         "Calculator",
		AppID:        "org.gnome.Calculator",
		Version:      "45.0",
		Branch:       "stable",
		Origin:       "flathub",
		Installation: "system",
	}
	b, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestListResultJSON(t *testing.T) {
	result := ListResult{
		Apps: []FlatpakInfo{
			{Name: "Calculator", AppID: "org.gnome.Calculator", Version: "45.0"},
		},
	}
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}
