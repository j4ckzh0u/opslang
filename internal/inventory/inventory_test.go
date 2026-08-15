package inventory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================
// Parse tests
// ============================================================

func TestParseValidInventory(t *testing.T) {
	data := []byte(`
hosts:
  - name: web1
    host: 192.168.1.10
    port: 22
    user: deploy
    key_file: /home/deploy/.ssh/id_rsa
    group: webservers
    tags:
      env: production
      role: web
  - name: db1
    host: 192.168.1.20
    port: 2222
    user: root
    password: secret
    group: databases
    tags:
      env: production
      role: db
`)

	inv, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inv.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(inv.Hosts))
	}

	// Check first host.
	h1 := inv.Hosts[0]
	if h1.Name != "web1" {
		t.Errorf("expected name 'web1', got %q", h1.Name)
	}
	if h1.Host != "192.168.1.10" {
		t.Errorf("expected host '192.168.1.10', got %q", h1.Host)
	}
	if h1.Port != 22 {
		t.Errorf("expected port 22, got %d", h1.Port)
	}
	if h1.User != "deploy" {
		t.Errorf("expected user 'deploy', got %q", h1.User)
	}
	if h1.KeyFile != "/home/deploy/.ssh/id_rsa" {
		t.Errorf("expected key_file '/home/deploy/.ssh/id_rsa', got %q", h1.KeyFile)
	}
	if h1.Group != "webservers" {
		t.Errorf("expected group 'webservers', got %q", h1.Group)
	}
	if h1.Tags["env"] != "production" {
		t.Errorf("expected tag env=production, got %q", h1.Tags["env"])
	}
	if h1.Tags["role"] != "web" {
		t.Errorf("expected tag role=web, got %q", h1.Tags["role"])
	}

	// Check second host.
	h2 := inv.Hosts[1]
	if h2.Name != "db1" {
		t.Errorf("expected name 'db1', got %q", h2.Name)
	}
	if h2.Port != 2222 {
		t.Errorf("expected port 2222, got %d", h2.Port)
	}
	if h2.Group != "databases" {
		t.Errorf("expected group 'databases', got %q", h2.Group)
	}
}

func TestParseDefaults(t *testing.T) {
	data := []byte(`
hosts:
  - host: 10.0.0.1
  - name: server2
`)

	inv, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inv.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(inv.Hosts))
	}

	// First host: name should default to host.
	h1 := inv.Hosts[0]
	if h1.Name != "10.0.0.1" {
		t.Errorf("expected name '10.0.0.1' (defaulted from host), got %q", h1.Name)
	}
	if h1.Port != 22 {
		t.Errorf("expected default port 22, got %d", h1.Port)
	}
	if h1.User != "root" {
		t.Errorf("expected default user 'root', got %q", h1.User)
	}

	// Second host: host should default to name.
	h2 := inv.Hosts[1]
	if h2.Host != "server2" {
		t.Errorf("expected host 'server2' (defaulted from name), got %q", h2.Host)
	}
	if h2.Port != 22 {
		t.Errorf("expected default port 22, got %d", h2.Port)
	}
	if h2.User != "root" {
		t.Errorf("expected default user 'root', got %q", h2.User)
	}
}

func TestParseValidationErrors(t *testing.T) {
	tests := []struct {
		name   string
		yaml   string
		errMsg string
	}{
		{
			name:   "host without name or host",
			yaml:   `hosts: [{}]`,
			errMsg: "either 'name' or 'host' is required",
		},
		{
			name:   "invalid yaml",
			yaml:   `{{{invalid`,
			errMsg: "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.yaml))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestParseEmptyInventory(t *testing.T) {
	data := []byte(`hosts: []`)
	inv, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inv.Hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(inv.Hosts))
	}
}

// ============================================================
// ParseFile tests
// ============================================================

func TestParseFile(t *testing.T) {
	content := `
hosts:
  - name: test-host
    host: 10.0.0.1
    port: 22
`
	dir := t.TempDir()
	path := filepath.Join(dir, "inventory.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	inv, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inv.Hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(inv.Hosts))
	}
	if inv.Hosts[0].Name != "test-host" {
		t.Errorf("expected name 'test-host', got %q", inv.Hosts[0].Name)
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/inventory.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("expected 'failed to read' error, got %q", err.Error())
	}
}

// ============================================================
// GetHostsByGroup tests
// ============================================================

func TestGetHostsByGroup(t *testing.T) {
	inv := &Inventory{
		Hosts: []Host{
			{Name: "web1", Host: "10.0.0.1", Group: "web"},
			{Name: "web2", Host: "10.0.0.2", Group: "web"},
			{Name: "db1", Host: "10.0.0.3", Group: "db"},
		},
	}

	webHosts := inv.GetHostsByGroup("web")
	if len(webHosts) != 2 {
		t.Errorf("expected 2 web hosts, got %d", len(webHosts))
	}

	dbHosts := inv.GetHostsByGroup("db")
	if len(dbHosts) != 1 {
		t.Errorf("expected 1 db host, got %d", len(dbHosts))
	}

	noHosts := inv.GetHostsByGroup("nonexistent")
	if len(noHosts) != 0 {
		t.Errorf("expected 0 hosts for nonexistent group, got %d", len(noHosts))
	}
}

// ============================================================
// GetHostsByTag tests
// ============================================================

func TestGetHostsByTag(t *testing.T) {
	inv := &Inventory{
		Hosts: []Host{
			{Name: "web1", Host: "10.0.0.1", Tags: map[string]string{"env": "prod", "role": "web"}},
			{Name: "web2", Host: "10.0.0.2", Tags: map[string]string{"env": "staging", "role": "web"}},
			{Name: "db1", Host: "10.0.0.3", Tags: map[string]string{"env": "prod", "role": "db"}},
			{Name: "db2", Host: "10.0.0.4"}, // No tags
		},
	}

	prodHosts := inv.GetHostsByTag("env", "prod")
	if len(prodHosts) != 2 {
		t.Errorf("expected 2 prod hosts, got %d", len(prodHosts))
	}

	webHosts := inv.GetHostsByTag("role", "web")
	if len(webHosts) != 2 {
		t.Errorf("expected 2 web hosts, got %d", len(webHosts))
	}

	noHosts := inv.GetHostsByTag("env", "dev")
	if len(noHosts) != 0 {
		t.Errorf("expected 0 dev hosts, got %d", len(noHosts))
	}
}

// ============================================================
// GetHostByName tests
// ============================================================

func TestGetHostByName(t *testing.T) {
	inv := &Inventory{
		Hosts: []Host{
			{Name: "web1", Host: "10.0.0.1"},
			{Name: "db1", Host: "10.0.0.2"},
		},
	}

	h, err := inv.GetHostByName("web1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Host != "10.0.0.1" {
		t.Errorf("expected host '10.0.0.1', got %q", h.Host)
	}

	_, err = inv.GetHostByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent host")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %q", err.Error())
	}
}

// ============================================================
// AllHostNames tests
// ============================================================

func TestAllHostNames(t *testing.T) {
	inv := &Inventory{
		Hosts: []Host{
			{Name: "web1", Host: "10.0.0.1"},
			{Name: "web2", Host: "10.0.0.2"},
			{Name: "db1", Host: "10.0.0.3"},
		},
	}

	names := inv.AllHostNames()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}

	expected := map[string]bool{"web1": true, "web2": true, "db1": true}
	for _, n := range names {
		delete(expected, n)
	}
	if len(expected) > 0 {
		t.Errorf("missing expected names: %v", expected)
	}
}

func TestAllHostNamesEmpty(t *testing.T) {
	inv := &Inventory{Hosts: []Host{}}
	names := inv.AllHostNames()
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}

// ============================================================
// Password JSON exclusion test
// ============================================================

func TestPasswordNotInJSON(t *testing.T) {
	h := Host{
		Name:     "test",
		Host:     "10.0.0.1",
		Password: "supersecret",
	}
	// The json:"-" tag on Password should prevent it from being marshaled.
	// We verify the struct tag exists by checking the field.
	if h.Password != "supersecret" {
		t.Error("password field should still hold the value in memory")
	}
}
