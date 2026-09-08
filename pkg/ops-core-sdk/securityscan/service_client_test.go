package securityscan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServiceBackendScan(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Error("missing authorization")
		}
		_, _ = w.Write([]byte(`{"backend":"service","status":"ok","database_version":"db-1"}`))
	}))
	defer server.Close()
	backend := &ServiceBackend{Config: &ServiceConfig{URL: server.URL, Token: "token", DatabaseVersion: "db-1"}, Client: server.Client()}
	report, err := backend.Scan(context.Background(), ScanRequest{Target: ScanTarget{ID: "host"}, Scanners: []Capability{CapabilityVulnerability}})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if report.Status != StatusOK || report.Target.ID != "host" {
		t.Fatalf("unexpected report: %#v", report)
	}
}
