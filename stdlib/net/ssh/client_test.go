package ssh

import (
	"testing"
)

func TestPoolKey(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "default port",
			cfg:  Config{Host: "10.0.0.1", User: "root"},
			want: "root@10.0.0.1:22",
		},
		{
			name: "custom port",
			cfg:  Config{Host: "10.0.0.1", Port: 2222, User: "admin"},
			want: "admin@10.0.0.1:2222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := poolKey(tt.cfg)
			if got != tt.want {
				t.Errorf("poolKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSSHConfig_NoAuth(t *testing.T) {
	cfg := Config{Host: "10.0.0.1", User: "root"}
	_, err := buildSSHConfig(cfg)
	if err == nil {
		t.Fatal("expected error when no auth method provided")
	}
}

func TestBuildSSHConfig_Password(t *testing.T) {
	cfg := Config{Host: "10.0.0.1", User: "root", Password: "secret"}
	sshCfg, err := buildSSHConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshCfg.User != "root" {
		t.Errorf("user = %q, want %q", sshCfg.User, "root")
	}
	if len(sshCfg.Auth) != 1 {
		t.Errorf("auth methods = %d, want 1", len(sshCfg.Auth))
	}
}

func TestBuildSSHConfig_BadKeyFile(t *testing.T) {
	cfg := Config{Host: "10.0.0.1", User: "root", KeyFile: "/nonexistent/key"}
	_, err := buildSSHConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent key file")
	}
}

func TestNewClient_MissingHost(t *testing.T) {
	_, err := NewClient(Config{User: "root", Password: "x"})
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestNewClient_MissingUser(t *testing.T) {
	_, err := NewClient(Config{Host: "10.0.0.1", Password: "x"})
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}

func TestPool_GetAndClose(t *testing.T) {
	p := NewPool()
	if p.Len() != 0 {
		t.Errorf("new pool len = %d, want 0", p.Len())
	}
	// Get with unreachable host should fail.
	cfg := Config{Host: "192.0.2.1", Port: 1, User: "root", Password: "x"}
	_, err := p.Get(cfg)
	if err == nil {
		t.Fatal("expected error connecting to unreachable host")
	}
	p.Close()
}

func TestClient_IsClosed(t *testing.T) {
	c := &Client{closed: false}
	if c.IsClosed() {
		t.Error("new client should not be closed")
	}
	c.closed = true
	if !c.IsClosed() {
		t.Error("closed client should report closed")
	}
}

func TestExecResult_Defaults(t *testing.T) {
	r := &ExecResult{}
	if r.Stdout != "" || r.Stderr != "" || r.ExitCode != 0 {
		t.Error("zero ExecResult should have empty strings and exit code 0")
	}
}
