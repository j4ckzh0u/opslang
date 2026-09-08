package runner

import (
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/securityscan"
)

func TestRegistryExternalScanner(t *testing.T) {
	for _, test := range []struct {
		name, body, wantError string
	}{
		{"complete", `{"status":"ok","coverage":{"complete":true},"database_version":"db-1","findings":[{"id":"TEST-1","package":"demo","installed_version":"1","fixed_version":"2"}]}`, ""},
		{"empty response", `{}`, "invalid scanner response status"},
		{"null response", `null`, "invalid scanner response status"},
		{"failed", `{"status":"failed","database_version":"db-1"}`, "did not complete"},
		{"incomplete", `{"status":"ok","database_version":"db-1"}`, "did not complete"},
		{"wrong database", `{"status":"ok","database_version":"db-2"}`, "version mismatch"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v1/security/scan" || r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("unexpected service request: %s %s", r.Method, r.URL.Path)
				}
				var request securityscan.ScanRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decode request: %v", err)
				}
				if request.Inventory == nil || len(request.Inventory.Packages) != 1 || request.Inventory.Packages[0].Name != "demo" || request.Target.ID != "test-host" {
					t.Errorf("missing inventory: %#v", request)
				}
				if _, err := w.Write([]byte(test.body)); err != nil {
					t.Errorf("write response: %v", err)
				}
			}))
			defer server.Close()
			ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
			scan, ok := NewRegistry().Get("security.scan")
			if !ok {
				t.Fatal("security.scan not registered")
			}
			result, err := scan(map[string]interface{}{
				"inventory": map[string]interface{}{"host": "test-host", "packages": []interface{}{map[string]interface{}{"name": "demo", "version": "1"}}},
				"scanners":  []string{"vuln"},
				"options": map[string]interface{}{"remote": securityscan.RemoteConfig{
					Backend: "external", URL: server.URL, Token: "test-token", CA: ca, RuleVersion: "db-1",
				}},
			})
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("error = %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			report, ok := result.(securityscan.ScanResult)
			if !ok || report.Status != securityscan.StatusOK || len(report.Findings) != 1 || report.Findings[0].ID != "TEST-1" || report.RuleSource.RuleVersion != "db-1" {
				t.Fatalf("unexpected scan result: %#v", result)
			}
		})
	}
}
