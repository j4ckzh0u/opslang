package ip_neighbor

import (
	"testing"
)

func TestList(t *testing.T) {
	r := List()
	// On non-Linux or without ip command, this may fail
	if r.Success {
		// Just verify it returns without error
		// In CI containers, there may be no neighbors
	}
}

func TestParseNeighbors(t *testing.T) {
	output := `192.168.1.1 dev eth0 lladdr aa:bb:cc:dd:ee:ff REACHABLE
192.168.1.2 dev eth0 lladdr 11:22:33:44:55:66 STALE router
fe80::1 dev eth0 lladdr aa:bb:cc:dd:ee:ff DELAY
10.0.0.1 dev eth0  FAILED`

	neighbors := parseNeighbors(output)
	if len(neighbors) != 4 {
		t.Fatalf("expected 4 neighbors, got %d", len(neighbors))
	}

	// Check first neighbor
	if neighbors[0].IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", neighbors[0].IP)
	}
	if neighbors[0].Dev != "eth0" {
		t.Errorf("expected dev eth0, got %s", neighbors[0].Dev)
	}
	if neighbors[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected MAC aa:bb:cc:dd:ee:ff, got %s", neighbors[0].MAC)
	}
	if neighbors[0].State != "REACHABLE" {
		t.Errorf("expected state REACHABLE, got %s", neighbors[0].State)
	}
	if neighbors[0].Router {
		t.Error("expected Router to be false")
	}

	// Check second neighbor (with router flag)
	if neighbors[1].IP != "192.168.1.2" {
		t.Errorf("expected IP 192.168.1.2, got %s", neighbors[1].IP)
	}
	if !neighbors[1].Router {
		t.Error("expected Router to be true")
	}

	// Check fourth neighbor (FAILED, no MAC)
	if neighbors[3].IP != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", neighbors[3].IP)
	}
	if neighbors[3].MAC != "" {
		t.Errorf("expected empty MAC, got %s", neighbors[3].MAC)
	}
	if neighbors[3].State != "FAILED" {
		t.Errorf("expected state FAILED, got %s", neighbors[3].State)
	}
}

func TestParseEmpty(t *testing.T) {
	neighbors := parseNeighbors("")
	if len(neighbors) != 0 {
		t.Errorf("expected 0 neighbors, got %d", len(neighbors))
	}
}

func TestParseNeighborLine(t *testing.T) {
	line := "192.168.1.100 dev ens192 lladdr 00:11:22:33:44:55 REACHABLE"
	neigh := parseNeighborLine(line)

	if neigh.IP != "192.168.1.100" {
		t.Errorf("expected IP 192.168.1.100, got %s", neigh.IP)
	}
	if neigh.Dev != "ens192" {
		t.Errorf("expected dev ens192, got %s", neigh.Dev)
	}
	if neigh.MAC != "00:11:22:33:44:55" {
		t.Errorf("expected MAC 00:11:22:33:44:55, got %s", neigh.MAC)
	}
	if neigh.State != "REACHABLE" {
		t.Errorf("expected state REACHABLE, got %s", neigh.State)
	}
}
