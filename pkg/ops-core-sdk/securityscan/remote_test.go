package securityscan

import (
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/software"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
)

func TestRemoteMatchHandlerRequiresTokenAndReturnsFindings(t *testing.T) {
	bundle, err := RuleBundleBytes(RuleBundle{
		FormatVersion: "1",
		RuleVersion:   "remote-test",
		Rules:         []vulnerability.Rule{{ID: "CVE-REMOTE", Package: "demo", FixedVersion: "2.0.0", Severity: "high"}},
	})
	if err != nil {
		t.Fatalf("RuleBundleBytes() error = %v", err)
	}
	handler, err := RemoteMatchHandler(bundle, "task-token")
	if err != nil {
		t.Fatalf("RemoteMatchHandler() error = %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	requestBody := `{"inventory":{"host":"node-1","packages":[{"name":"demo","version":"1.0.0","manager":"apt"}]}}`
	request, err := http.NewRequest(http.MethodPost, server.URL+remoteMatchPath, strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	request.Header.Set("Authorization", "Bearer task-token")
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	var payload remoteMatchResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Findings) != 1 || payload.Findings[0].ID != "CVE-REMOTE" {
		t.Fatalf("findings = %#v, want one CVE-REMOTE finding", payload.Findings)
	}
	if payload.RuleSource.RuleVersion != "remote-test" {
		t.Fatalf("rule version = %q, want remote-test", payload.RuleSource.RuleVersion)
	}
}

func TestRemoteMatchHandlerRejectsWrongToken(t *testing.T) {
	bundle, err := RuleBundleBytes(RuleBundle{FormatVersion: "1", RuleVersion: "remote-test", Rules: []vulnerability.Rule{{ID: "CVE-1", Package: "demo"}}})
	if err != nil {
		t.Fatalf("RuleBundleBytes() error = %v", err)
	}
	handler, err := RemoteMatchHandler(bundle, "task-token")
	if err != nil {
		t.Fatalf("RemoteMatchHandler() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, remoteMatchPath, strings.NewReader(`{}`))
	request.Header.Set("Authorization", "Bearer wrong-token")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestRemoteMatchHandlerRejectsTrailingJSON(t *testing.T) {
	bundle, err := RuleBundleBytes(RuleBundle{FormatVersion: "1", RuleVersion: "remote-test", Rules: []vulnerability.Rule{{ID: "CVE-1", Package: "demo"}}})
	if err != nil {
		t.Fatalf("RuleBundleBytes() error = %v", err)
	}
	handler, err := RemoteMatchHandler(bundle, "task-token")
	if err != nil {
		t.Fatalf("RemoteMatchHandler() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, remoteMatchPath, strings.NewReader(`{"inventory":{}} {}`))
	request.Header.Set("Authorization", "Bearer task-token")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestQueryRemoteVulnerabilitiesRequiresHTTPS(t *testing.T) {
	_, _, err := queryRemoteVulnerabilities(t.Context(), software.InventoryResult{}, &RemoteConfig{URL: "http://controller.example", Token: "task-token"})
	if err == nil || !strings.Contains(err.Error(), "HTTPS") {
		t.Fatalf("error = %v, want HTTPS validation error", err)
	}
}

func TestQueryRemoteVulnerabilitiesOverTLS(t *testing.T) {
	bundle, err := RuleBundleBytes(RuleBundle{
		FormatVersion: "1",
		RuleVersion:   "remote-tls",
		Rules:         []vulnerability.Rule{{ID: "CVE-TLS", Package: "demo", FixedVersion: "2.0.0", Severity: "high"}},
	})
	if err != nil {
		t.Fatalf("RuleBundleBytes() error = %v", err)
	}
	_, source, err := ValidateRuleBundle(bundle)
	if err != nil {
		t.Fatalf("ValidateRuleBundle() error = %v", err)
	}
	handler, err := RemoteMatchHandler(bundle, "task-token")
	if err != nil {
		t.Fatalf("RemoteMatchHandler() error = %v", err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	certificate := server.Certificate()
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})

	inventory := software.InventoryResult{Host: "node-1", Packages: []software.Package{{Name: "demo", Version: "1.0.0", Manager: "apt"}}}
	findings, actualSource, err := queryRemoteVulnerabilities(t.Context(), inventory, &RemoteConfig{
		URL:         server.URL,
		Token:       "task-token",
		RuleVersion: source.RuleVersion,
		RuleSHA256:  source.SHA256,
		CA:          caPEM,
	})
	if err != nil {
		t.Fatalf("queryRemoteVulnerabilities() error = %v", err)
	}
	if len(findings) != 1 || findings[0].ID != "CVE-TLS" {
		t.Fatalf("findings = %#v, want one CVE-TLS finding", findings)
	}
	if actualSource != source {
		t.Fatalf("rule source = %#v, want %#v", actualSource, source)
	}
}

func TestRemoteConfigJSONCarriesCA(t *testing.T) {
	config := RemoteConfig{URL: "https://controller.example", Token: "task-token", CA: []byte("ca-data")}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !strings.Contains(string(data), `"ca":"Y2EtZGF0YQ=="`) {
		t.Fatalf("JSON = %s, want encoded CA", data)
	}
}

func TestValidateRemoteConfigRejectsInvalidDigestAndCA(t *testing.T) {
	tests := []struct {
		name   string
		config RemoteConfig
		want   string
	}{
		{name: "digest", config: RemoteConfig{URL: "https://controller.example", Token: "task-token", RuleSHA256: "bad"}, want: "64-character"},
		{name: "CA", config: RemoteConfig{URL: "https://controller.example", Token: "task-token", CA: []byte("bad")}, want: "no certificates"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateRemoteConfig(&test.config)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}
