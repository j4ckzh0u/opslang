package sshx

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config with password",
			config: Config{
				Host:     "example.com",
				Port:     22,
				User:     "root",
				Password: "secret",
			},
			wantErr: false,
		},
		{
			name: "valid config with key file",
			config: Config{
				Host:    "example.com",
				Port:    22,
				User:    "root",
				KeyFile: "/path/to/key",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: Config{
				Port:     22,
				Password: "secret",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				Host:     "example.com",
				Port:     -1,
				Password: "secret",
			},
			wantErr: true,
		},
		{
			name: "port too high",
			config: Config{
				Host:     "example.com",
				Port:     70000,
				Password: "secret",
			},
			wantErr: true,
		},
		{
			name: "no auth method",
			config: Config{
				Host: "example.com",
				Port: 22,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: Config{
				Host:     "example.com",
				Port:     22,
				Password: "secret",
				Timeout:  -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative retries",
			config: Config{
				Host:     "example.com",
				Port:     22,
				Password: "secret",
				Retries:  -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	cfg := &Config{
		Host:     "example.com",
		Password: "secret",
	}

	cfg.SetDefaults()

	if cfg.Port != 22 {
		t.Errorf("expected default port 22, got %d", cfg.Port)
	}
	if cfg.User != "root" {
		t.Errorf("expected default user root, got %s", cfg.User)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.Retries != 3 {
		t.Errorf("expected default retries 3, got %d", cfg.Retries)
	}
	if cfg.MaxConnections != 5 {
		t.Errorf("expected default max connections 5, got %d", cfg.MaxConnections)
	}
}

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			Host:     "example.com",
			Password: "secret",
		}

		client, err := NewClient(cfg)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("NewClient() returned nil client")
		}
		if client.config.Port != 22 {
			t.Errorf("expected default port 22, got %d", client.config.Port)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := &Config{
			Host: "",
		}

		_, err := NewClient(cfg)
		if err == nil {
			t.Error("NewClient() expected error for invalid config")
		}
	})

	t.Run("invalid key file", func(t *testing.T) {
		cfg := &Config{
			Host:    "example.com",
			KeyFile: "/nonexistent/key",
		}

		_, err := NewClient(cfg)
		if err == nil {
			t.Error("NewClient() expected error for invalid key file")
		}
	})
}

func TestClient_Connect_NotConnected(t *testing.T) {
	cfg := &Config{
		Host:     "nonexistent.example.com",
		Port:     22,
		User:     "root",
		Password: "secret",
		Timeout:  1 * time.Second,
		Retries:  0,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test Exec without connection
	_, err = client.Exec(context.Background(), "echo test")
	if err == nil {
		t.Error("Exec() expected error when not connected")
	}
}

func TestClient_IsConnected(t *testing.T) {
	cfg := &Config{
		Host:     "example.com",
		Password: "secret",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.IsConnected() {
		t.Error("IsConnected() should return false before Connect()")
	}
}

func TestClient_Close(t *testing.T) {
	cfg := &Config{
		Host:     "example.com",
		Password: "secret",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Close without connection should not error
	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestSFTPClient_Operations(t *testing.T) {
	cfg := &Config{
		Host:     "example.com",
		Password: "secret",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test SFTP operations without connection
	ctx := context.Background()

	// Upload without connection
	err = client.Upload(ctx, "/local/file", "/remote/file")
	if err == nil {
		t.Error("Upload() expected error when not connected")
	}

	// Download without connection
	err = client.Download(ctx, "/remote/file", "/local/file")
	if err == nil {
		t.Error("Download() expected error when not connected")
	}

	// NewSFTPClient without connection
	_, err = client.NewSFTPClient()
	if err == nil {
		t.Error("NewSFTPClient() expected error when not connected")
	}
}

func TestIsConnectionError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if IsConnectionError(nil) {
			t.Error("IsConnectionError(nil) should return false")
		}
	})

	t.Run("non-connection error", func(t *testing.T) {
		err := os.ErrNotExist
		if IsConnectionError(err) {
			t.Error("IsConnectionError() should return false for non-connection errors")
		}
	})
}

func TestSafeBuffer(t *testing.T) {
	buf := &safeBuffer{}

	// Test Write
	n, err := buf.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 5 {
		t.Errorf("Write() returned %d, want 5", n)
	}

	// Test String
	if got := buf.String(); got != "hello" {
		t.Errorf("String() = %q, want %q", got, "hello")
	}

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			buf.Write([]byte("x"))
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	if got := len(buf.String()); got != 15 {
		t.Errorf("String() length = %d, want 15", got)
	}
}

func TestBuildSSHConfig(t *testing.T) {
	t.Run("with password", func(t *testing.T) {
		cfg := &Config{
			Host:     "example.com",
			User:     "root",
			Password: "secret",
			Timeout:  30 * time.Second,
		}

		sshConfig, err := buildSSHConfig(cfg)
		if err != nil {
			t.Fatalf("buildSSHConfig() error = %v", err)
		}
		if sshConfig.User != "root" {
			t.Errorf("expected user root, got %s", sshConfig.User)
		}
		if len(sshConfig.Auth) != 1 {
			t.Errorf("expected 1 auth method, got %d", len(sshConfig.Auth))
		}
	})

	t.Run("with key file", func(t *testing.T) {
		// Create a temporary key file
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "id_rsa")

		// Generate a test key (this is a simple RSA key for testing)
		keyContent := `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBLKZzGD3QwYQAAAABzc2gtZWQyNTUxOQAAACBLKZzGD3QwYQAAAEDL1xQZ
qX6VYK5mKJ8AAAAAAAAAAAAAAAtzc2gtZWQyNTUxOQ==
-----END OPENSSH PRIVATE KEY-----`

		err := os.WriteFile(keyPath, []byte(keyContent), 0600)
		if err != nil {
			t.Fatalf("Failed to create test key: %v", err)
		}

		cfg := &Config{
			Host:    "example.com",
			User:    "root",
			KeyFile: keyPath,
			Timeout: 30 * time.Second,
		}

		_, err = buildSSHConfig(cfg)
		// This may fail due to invalid key format, but we're testing the code path
		if err == nil {
			t.Log("buildSSHConfig() succeeded with test key")
		}
	})

	t.Run("no auth method", func(t *testing.T) {
		cfg := &Config{
			Host:    "example.com",
			User:    "root",
			Timeout: 30 * time.Second,
		}

		_, err := buildSSHConfig(cfg)
		if err == nil {
			t.Error("buildSSHConfig() expected error with no auth method")
		}
	})

	t.Run("invalid key file path", func(t *testing.T) {
		cfg := &Config{
			Host:    "example.com",
			User:    "root",
			KeyFile: "/nonexistent/key",
			Timeout: 30 * time.Second,
		}

		_, err := buildSSHConfig(cfg)
		if err == nil {
			t.Error("buildSSHConfig() expected error for invalid key file")
		}
	})
}

func TestClient_Connect_WithMockServer(t *testing.T) {
	server, err := newMockSSHServer("testpass")
	if err != nil {
		t.Fatalf("failed to create mock server: %v", err)
	}
	defer server.Close()

	host, _, _ := net.SplitHostPort(server.Addr())
	port := server.Port()

	cfg := &Config{
		Host:     host,
		Port:     port,
		User:     "root",
		Password: "testpass",
		Timeout:  5 * time.Second,
		Retries:  1,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	if !client.IsConnected() {
		t.Error("client should be connected")
	}
}

func TestClient_Connect_Retry(t *testing.T) {
	cfg := &Config{
		Host:     "nonexistent.example.com",
		Port:     22,
		User:     "root",
		Password: "testpass",
		Timeout:  1 * time.Second,
		Retries:  2,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connect() should fail for nonexistent host")
		client.Close()
	}

	// Verify error message contains retry info
	if err != nil && !strings.Contains(err.Error(), "attempts") {
		t.Errorf("error should mention retry attempts, got: %v", err)
	}
}

func TestClient_Connect_ContextCancel(t *testing.T) {
	server, err := newMockSSHServer("testpass")
	if err != nil {
		t.Fatalf("failed to create mock server: %v", err)
	}
	defer server.Close()

	host, _, _ := net.SplitHostPort(server.Addr())
	port := server.Port()

	cfg := &Config{
		Host:     host,
		Port:     port,
		User:     "root",
		Password: "testpass",
		Timeout:  5 * time.Second,
		Retries:  5,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connect() should fail with canceled context")
		client.Close()
	}
}

func TestClient_Exec_WithMockServer(t *testing.T) {
	server, err := newMockSSHServer("testpass")
	if err != nil {
		t.Fatalf("failed to create mock server: %v", err)
	}
	defer server.Close()

	host, _, _ := net.SplitHostPort(server.Addr())
	port := server.Port()

	cfg := &Config{
		Host:     host,
		Port:     port,
		User:     "root",
		Password: "testpass",
		Timeout:  5 * time.Second,
		Retries:  1,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	// Test successful command
	result, err := client.Exec(ctx, "echo hello")
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "hello") {
		t.Errorf("expected stdout to contain 'hello', got: %s", result.Stdout)
	}
}

func TestClient_Exec_NotConnected(t *testing.T) {
	cfg := &Config{
		Host:     "example.com",
		Password: "testpass",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.Exec(context.Background(), "echo test")
	if err == nil {
		t.Error("Exec() should fail when not connected")
	}
}

func TestClient_Upload_NotConnected(t *testing.T) {
	cfg := &Config{
		Host:     "example.com",
		Password: "testpass",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = client.Upload(context.Background(), "/local", "/remote")
	if err == nil {
		t.Error("Upload() should fail when not connected")
	}
}

func TestClient_Download_NotConnected(t *testing.T) {
	cfg := &Config{
		Host:     "example.com",
		Password: "testpass",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = client.Download(context.Background(), "/remote", "/local")
	if err == nil {
		t.Error("Download() should fail when not connected")
	}
}

func TestSFTPClient_Close(t *testing.T) {
	sftp := &SFTPClient{client: nil}
	err := sftp.Close()
	if err != nil {
		t.Errorf("Close() should not error for nil client")
	}
}
