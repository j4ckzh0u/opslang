package securityscan

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
)

func TestValidateScanners(t *testing.T) {
	tests := []struct {
		name     string
		scanners []Scanner
		wantErr  bool
	}{
		{name: "supported", scanners: []Scanner{ScannerVulnerability, ScannerSBOM}},
		{name: "empty", scanners: nil, wantErr: true},
		{name: "unsupported", scanners: []Scanner{ScannerSecret}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateScanners(test.scanners)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateScanners() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateOptionsRejectsInvalidLimitsSeverityAndFormat(t *testing.T) {
	tests := []Options{
		{Scanners: []Scanner{ScannerSBOM}, MaxFileSize: -1},
		{Scanners: []Scanner{ScannerSBOM}, MaxDepth: -1},
		{Scanners: []Scanner{ScannerSBOM}, Timeout: -1},
		{Scanners: []Scanner{ScannerSBOM}, Severity: "urgent"},
		{Scanners: []Scanner{ScannerSBOM}, Format: "xml"},
	}
	for _, options := range tests {
		if err := ValidateOptions(options); err == nil {
			t.Fatalf("ValidateOptions(%#v) returned nil error", options)
		}
	}
}

func TestRuleBundleBytesAndValidation(t *testing.T) {
	bundle := RuleBundle{
		FormatVersion: "1",
		RuleVersion:   "2026-09-06",
		GeneratedAt:   time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC),
		Rules:         []vulnerability.Rule{{ID: "CVE-2026-0001", Package: "openssl", FixedVersion: "3.0.1"}},
	}
	data, err := RuleBundleBytes(bundle)
	if err != nil {
		t.Fatalf("RuleBundleBytes() error = %v", err)
	}
	decoded, source, err := ValidateRuleBundle(data)
	if err != nil {
		t.Fatalf("ValidateRuleBundle() error = %v", err)
	}
	if decoded.Rules[0].ID != "CVE-2026-0001" || source.SHA256 == "" {
		t.Fatalf("validated bundle = %#v, source = %#v", decoded, source)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	payload["sha256"] = "bad"
	corrupted, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if _, _, err := ValidateRuleBundle(corrupted); err == nil {
		t.Fatal("ValidateRuleBundle() accepted a mismatched digest")
	}
}

func TestValidateRuleBundleRejectsNilAndMalformedInput(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("{"), []byte(`{"format_version":"2"}`)} {
		if _, _, err := ValidateRuleBundle(data); err == nil {
			t.Fatalf("ValidateRuleBundle(%q) returned nil error", data)
		}
	}
}

func TestScanResultInitializesCollections(t *testing.T) {
	result := NewScanResult(TargetInfo{Type: "filesystem", ID: "tmp"}, []Scanner{ScannerSBOM})
	if result.Findings == nil || result.Components == nil || result.Errors == nil {
		t.Fatalf("NewScanResult() returned nil collections: %#v", result)
	}
	result.Finish()
	if result.FinishedAt.IsZero() {
		t.Fatal("Finish() did not set FinishedAt")
	}
	var nilResult *ScanResult
	nilResult.Finish()
}

func TestStableSortAndComponentID(t *testing.T) {
	findings := []VulnerabilityFinding{
		{ID: "B", Package: "z"},
		{ID: "A", Package: "z"},
	}
	StableSortFindings(findings)
	if findings[0].ID != "A" {
		t.Fatalf("StableSortFindings() first ID = %q", findings[0].ID)
	}
	components := []SBOMComponent{
		{Ecosystem: "npm", Name: "z", Version: "1"},
		{Ecosystem: "go", Name: "a", Version: "1"},
	}
	StableSortComponents(components)
	if components[0].Ecosystem != "go" {
		t.Fatalf("StableSortComponents() first ecosystem = %q", components[0].Ecosystem)
	}
	if got := ComponentID(" Go ", "example", "1.0"); got != "go:example:1.0" {
		t.Fatalf("ComponentID() = %q", got)
	}
}
