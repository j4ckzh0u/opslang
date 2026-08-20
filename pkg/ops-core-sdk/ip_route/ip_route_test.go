package ip_route

import (
	"testing"
)

func TestParseRouteLine(t *testing.T) {
	tests := []struct {
		line string
		want Route
	}{
		{
			line: "default via 192.168.1.1 dev eth0",
			want: Route{Destination: "default", Gateway: "192.168.1.1", Dev: "eth0"},
		},
		{
			line: "10.0.0.0/8 via 10.0.0.1 dev eth0 metric 100",
			want: Route{Destination: "10.0.0.0/8", Gateway: "10.0.0.1", Dev: "eth0", Metric: 100},
		},
		{
			line: "192.168.1.0/24 dev eth0 scope link",
			want: Route{Destination: "192.168.1.0/24", Dev: "eth0", Scope: "link"},
		},
		{
			line: "172.16.0.0/12 dev tun0 table 100",
			want: Route{Destination: "172.16.0.0/12", Dev: "tun0", Table: "100"},
		},
	}

	for _, tt := range tests {
		got := parseRouteLine(tt.line)
		if got.Destination != tt.want.Destination {
			t.Errorf("parseRouteLine(%q) destination = %q, want %q", tt.line, got.Destination, tt.want.Destination)
		}
		if got.Gateway != tt.want.Gateway {
			t.Errorf("parseRouteLine(%q) gateway = %q, want %q", tt.line, got.Gateway, tt.want.Gateway)
		}
		if got.Dev != tt.want.Dev {
			t.Errorf("parseRouteLine(%q) dev = %q, want %q", tt.line, got.Dev, tt.want.Dev)
		}
		if got.Metric != tt.want.Metric {
			t.Errorf("parseRouteLine(%q) metric = %d, want %d", tt.line, got.Metric, tt.want.Metric)
		}
	}
}

func TestParseRoutes(t *testing.T) {
	output := `default via 192.168.1.1 dev eth0
10.0.0.0/8 via 10.0.0.1 dev eth0
192.168.1.0/24 dev eth0 scope link`

	routes := parseRoutes(output)

	if len(routes) != 3 {
		t.Fatalf("expected 3 routes, got %d", len(routes))
	}
	if routes[0].Destination != "default" {
		t.Errorf("first route should be default, got %s", routes[0].Destination)
	}
	if routes[1].Gateway != "10.0.0.1" {
		t.Errorf("second route gateway should be 10.0.0.1, got %s", routes[1].Gateway)
	}
}

func TestAddValidation(t *testing.T) {
	result := Add(AddConfig{})
	if result.Success {
		t.Error("Add() should fail with empty destination")
	}
}

func TestDeleteValidation(t *testing.T) {
	result := Delete("", "")
	if result.Success {
		t.Error("Delete() should fail with empty destination")
	}
}

func TestListRequiresIP(t *testing.T) {
	// This test may fail if ip command is not available (e.g., on macOS)
	// but should not panic
	result := List()
	// Just verify it returns without panic
	_ = result.Duration
}
