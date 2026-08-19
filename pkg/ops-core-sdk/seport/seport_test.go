package seport

import (
	"testing"
)

func TestListResultJSON(t *testing.T) {
	r := ListResult{
		Ports: []PortEntry{
			{SELinuxPortType: "ssh_port_t", Protocol: "tcp", PortNumber: "22"},
		},
	}
	if len(r.Ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(r.Ports))
	}
	if r.Ports[0].SELinuxPortType != "ssh_port_t" {
		t.Errorf("expected ssh_port_t, got %s", r.Ports[0].SELinuxPortType)
	}
}

func TestAddValidation(t *testing.T) {
	_, err := Add("", "tcp", "8080")
	if err == nil {
		t.Error("expected error for empty seport_type")
	}

	_, err = Add("http_port_t", "", "8080")
	if err == nil {
		t.Error("expected error for empty protocol")
	}

	_, err = Add("http_port_t", "tcp", "")
	if err == nil {
		t.Error("expected error for empty port")
	}
}

func TestRemoveValidation(t *testing.T) {
	_, err := Remove("", "8080")
	if err == nil {
		t.Error("expected error for empty protocol")
	}

	_, err = Remove("tcp", "")
	if err == nil {
		t.Error("expected error for empty port")
	}
}

func TestGetValidation(t *testing.T) {
	_, err := Get("", "8080")
	if err == nil {
		t.Error("expected error for empty protocol")
	}

	_, err = Get("tcp", "")
	if err == nil {
		t.Error("expected error for empty port")
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Changed: true, Message: "test"}
	if !r.Changed {
		t.Error("expected Changed=true")
	}
	if r.Message != "test" {
		t.Errorf("expected 'test', got %s", r.Message)
	}
}
