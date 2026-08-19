package firewalld_zone

import (
	"testing"
)

func TestAddService_RequiresZoneAndService(t *testing.T) {
	_, err := AddService("", "http")
	if err == nil {
		t.Error("expected error for empty zone")
	}
	_, err = AddService("public", "")
	if err == nil {
		t.Error("expected error for empty service")
	}
}

func TestRemoveService_RequiresZoneAndService(t *testing.T) {
	_, err := RemoveService("", "http")
	if err == nil {
		t.Error("expected error for empty zone")
	}
	_, err = RemoveService("public", "")
	if err == nil {
		t.Error("expected error for empty service")
	}
}

func TestAddPort_RequiresZoneAndPort(t *testing.T) {
	_, err := AddPort("", "8080/tcp")
	if err == nil {
		t.Error("expected error for empty zone")
	}
	_, err = AddPort("public", "")
	if err == nil {
		t.Error("expected error for empty port")
	}
}

func TestRemovePort_RequiresZoneAndPort(t *testing.T) {
	_, err := RemovePort("", "8080/tcp")
	if err == nil {
		t.Error("expected error for empty zone")
	}
	_, err = RemovePort("public", "")
	if err == nil {
		t.Error("expected error for empty port")
	}
}

func TestAddRichRule_RequiresZoneAndRule(t *testing.T) {
	_, err := AddRichRule("", "rule")
	if err == nil {
		t.Error("expected error for empty zone")
	}
	_, err = AddRichRule("public", "")
	if err == nil {
		t.Error("expected error for empty rule")
	}
}

func TestRemoveRichRule_RequiresZoneAndRule(t *testing.T) {
	_, err := RemoveRichRule("", "rule")
	if err == nil {
		t.Error("expected error for empty zone")
	}
	_, err = RemoveRichRule("public", "")
	if err == nil {
		t.Error("expected error for empty rule")
	}
}

func TestSetDefaultZone_RequiresZone(t *testing.T) {
	_, err := SetDefaultZone("")
	if err == nil {
		t.Error("expected error for empty zone")
	}
}

func TestAddZone_RequiresZone(t *testing.T) {
	_, err := AddZone("")
	if err == nil {
		t.Error("expected error for empty zone")
	}
}

func TestRemoveZone_RequiresZone(t *testing.T) {
	_, err := RemoveZone("")
	if err == nil {
		t.Error("expected error for empty zone")
	}
}

func TestInfo_RequiresZone(t *testing.T) {
	_, err := Info("")
	if err == nil {
		t.Error("expected error for empty zone")
	}
}

func TestActionResultFields(t *testing.T) {
	r := ActionResult{
		Zone:    "public",
		Action:  "add_service",
		Changed: true,
		Message: "service http added",
	}
	if r.Zone != "public" {
		t.Error("zone mismatch")
	}
	if r.Action != "add_service" {
		t.Error("action mismatch")
	}
}

func TestZoneInfoFields(t *testing.T) {
	info := ZoneInfo{
		Name:     "public",
		Default:  true,
		Services: []string{"ssh", "http"},
		Ports:    []string{"8080/tcp"},
	}
	if info.Name != "public" {
		t.Error("name mismatch")
	}
	if !info.Default {
		t.Error("default should be true")
	}
	if len(info.Services) != 2 {
		t.Error("expected 2 services")
	}
}

func TestListResultFields(t *testing.T) {
	r := ListResult{
		Zones: []string{"public", "internal", "dmz"},
		Count: 3,
	}
	if r.Count != 3 {
		t.Error("count mismatch")
	}
}
