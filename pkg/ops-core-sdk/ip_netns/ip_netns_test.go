package ip_netns

import (
	"testing"
)

func TestList(t *testing.T) {
	r := List()
	// On non-Linux or without ip command, this may fail
	if r.Success {
		// Just verify it returns without error
		// In CI containers, there may be no namespaces
	}
}

func TestGetNonExistent(t *testing.T) {
	r := Get("nonexistent_ns_xyz")
	if !r.Success {
		t.Error("expected success even for non-existent namespace")
	}
	if len(r.Namespaces) == 0 {
		t.Error("expected namespace info")
	}
	if r.Namespaces[0].Exists {
		t.Error("expected Exists to be false for non-existent namespace")
	}
}

func TestParseNamespaces(t *testing.T) {
	output := `ns1 (id: 0)
ns2 (id: 1)
ns3`

	namespaces := parseNamespaces(output)
	if len(namespaces) != 3 {
		t.Fatalf("expected 3 namespaces, got %d", len(namespaces))
	}

	// Check ns1
	if namespaces[0].Name != "ns1" {
		t.Errorf("expected name ns1, got %s", namespaces[0].Name)
	}
	if namespaces[0].ID != "0" {
		t.Errorf("expected ID 0, got %s", namespaces[0].ID)
	}
	if !namespaces[0].Exists {
		t.Error("expected Exists to be true")
	}

	// Check ns2
	if namespaces[1].Name != "ns2" {
		t.Errorf("expected name ns2, got %s", namespaces[1].Name)
	}
	if namespaces[1].ID != "1" {
		t.Errorf("expected ID 1, got %s", namespaces[1].ID)
	}

	// Check ns3 (no ID)
	if namespaces[2].Name != "ns3" {
		t.Errorf("expected name ns3, got %s", namespaces[2].Name)
	}
	if namespaces[2].ID != "" {
		t.Errorf("expected empty ID, got %s", namespaces[2].ID)
	}
}

func TestParseEmpty(t *testing.T) {
	namespaces := parseNamespaces("")
	if len(namespaces) != 0 {
		t.Errorf("expected 0 namespaces, got %d", len(namespaces))
	}
}
