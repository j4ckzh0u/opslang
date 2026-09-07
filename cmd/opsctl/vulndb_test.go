package main

import (
	"encoding/pem"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadRemoteVulnerabilityConfigDisabled(t *testing.T) {
	config, err := loadRemoteVulnerabilityConfig("", "", "", "", "", 30*time.Second)
	if err != nil {
		t.Fatalf("loadRemoteVulnerabilityConfig() error = %v", err)
	}
	if config != nil {
		t.Fatalf("config = %#v, want nil", config)
	}
}

func TestLoadRemoteVulnerabilityConfigReadsCA(t *testing.T) {
	server := httptest.NewTLSServer(nil)
	defer server.Close()
	ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	caPath := filepath.Join(t.TempDir(), "controller-ca.pem")
	if err := os.WriteFile(caPath, ca, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	config, err := loadRemoteVulnerabilityConfig("https://controller.example", "task-token", "rules-v1", strings.Repeat("a", 64), caPath, time.Second)
	if err != nil {
		t.Fatalf("loadRemoteVulnerabilityConfig() error = %v", err)
	}
	if string(config.CA) != string(ca) || config.Timeout != time.Second {
		t.Fatalf("config = %#v, want CA and timeout", config)
	}
}

func TestLoadRemoteVulnerabilityConfigRejectsPartialSettings(t *testing.T) {
	config, err := loadRemoteVulnerabilityConfig("https://controller.example", "", "", "", "", time.Second)
	if err == nil || !strings.Contains(err.Error(), "token") {
		t.Fatalf("config = %#v, error = %v, want token validation error", config, err)
	}
}

func TestHasRemoteVulnerabilitySettings(t *testing.T) {
	if hasRemoteVulnerabilitySettings("", "", "", "", "") {
		t.Fatal("empty settings reported as enabled")
	}
	if !hasRemoteVulnerabilitySettings("https://controller.example", "", "", "", "") {
		t.Fatal("endpoint did not enable remote settings")
	}
}
