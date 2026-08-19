package nmcli

import (
	"encoding/json"
	"testing"
)

func TestAdd_EmptyArgs(t *testing.T) {
	_, err := Add("", "ethernet", nil)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	_, err = Add("test", "", nil)
	if err == nil {
		t.Fatal("expected error for empty type")
	}
}

func TestModify_EmptyName(t *testing.T) {
	_, err := Modify("", nil)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDelete_EmptyName(t *testing.T) {
	_, err := Delete("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUp_EmptyName(t *testing.T) {
	_, err := Up("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDown_EmptyName(t *testing.T) {
	_, err := Down("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestShow_EmptyName(t *testing.T) {
	_, err := Show("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Name: "eth0", Changed: true, Success: true}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestConnectionInfoJSON(t *testing.T) {
	conn := ConnectionInfo{
		Name:   "eth0",
		UUID:   "12345",
		Type:   "ethernet",
		Device: "eth0",
		State:  "connected",
	}
	b, err := json.Marshal(conn)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestDeviceInfoJSON(t *testing.T) {
	dev := DeviceInfo{
		Device:     "eth0",
		Type:       "ethernet",
		State:      "connected",
		Connection: "eth0",
	}
	b, err := json.Marshal(dev)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestListResultJSON(t *testing.T) {
	result := ListResult{
		Connections: []ConnectionInfo{
			{Name: "eth0", UUID: "12345", Type: "ethernet"},
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

func TestDeviceListResultJSON(t *testing.T) {
	result := DeviceListResult{
		Devices: []DeviceInfo{
			{Device: "eth0", Type: "ethernet"},
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

func TestConnectionSettingsJSON(t *testing.T) {
	settings := ConnectionSettings{
		Settings: map[string]map[string]interface{}{
			"connection": {
				"id":   "eth0",
				"type": "ethernet",
			},
		},
	}
	b, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}
