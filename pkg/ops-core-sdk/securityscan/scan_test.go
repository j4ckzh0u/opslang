package securityscan

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/software"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
)

func TestScanInventoryMatchesRulesAndBuildsSBOM(t *testing.T) {
	ruleData, err := RuleBundleBytes(RuleBundle{
		FormatVersion: "1",
		RuleVersion:   "test",
		Rules:         []vulnerability.Rule{{ID: "CVE-1", Package: "openssl", FixedVersion: "3.0.1", Severity: "high"}},
	})
	if err != nil {
		t.Fatalf("RuleBundleBytes() error = %v", err)
	}
	result, err := ScanInventory(software.InventoryResult{
		Host:     "host-a",
		Packages: []software.Package{{Name: "openssl", Version: "3.0.0", Manager: "apt"}},
	}, Options{Scanners: []Scanner{ScannerVulnerability, ScannerSBOM}, RuleBundle: ruleData})
	if err != nil {
		t.Fatalf("ScanInventory() error = %v", err)
	}
	if len(result.Findings) != 1 || result.Findings[0].ID != "CVE-1" {
		t.Fatalf("findings = %#v", result.Findings)
	}
	if len(result.Components) != 1 || result.Components[0].Name != "openssl" {
		t.Fatalf("components = %#v", result.Components)
	}
}

func TestScanInventoryAppliesIgnoreRulesAndSeverityGate(t *testing.T) {
	ruleData, err := RuleBundleBytes(RuleBundle{
		FormatVersion: "1",
		RuleVersion:   "test",
		Rules: []vulnerability.Rule{
			{ID: "CVE-LOW", Package: "demo", FixedVersion: "2.0.0", Severity: "low"},
			{ID: "CVE-HIGH", Package: "demo", FixedVersion: "2.0.0", Severity: "high"},
		},
	})
	if err != nil {
		t.Fatalf("RuleBundleBytes() error = %v", err)
	}
	inventory := software.InventoryResult{Packages: []software.Package{{Name: "demo", Version: "1.0.0", InstalledFiles: []string{"/usr/bin/demo"}}}}
	result, err := ScanInventory(inventory, Options{Scanners: []Scanner{ScannerVulnerability}, RuleBundle: ruleData, Severity: "medium"})
	if err != nil {
		t.Fatalf("ScanInventory() error = %v", err)
	}
	if result.Status != StatusFailed || len(result.Findings) != 1 || result.Findings[0].ID != "CVE-HIGH" || result.Findings[0].Path != "/usr/bin/demo" {
		t.Fatalf("gated result = %#v", result)
	}
	result, err = ScanInventory(inventory, Options{Scanners: []Scanner{ScannerVulnerability}, RuleBundle: ruleData, Severity: "medium", IgnoreRules: []IgnoreRule{{ID: "CVE-HIGH"}}})
	if err != nil {
		t.Fatalf("ignored ScanInventory() error = %v", err)
	}
	if result.Status != StatusOK || len(result.Findings) != 0 {
		t.Fatalf("ignored result = %#v", result)
	}
}

func TestScanInventoryValueAcceptsRuleBundleObject(t *testing.T) {
	ruleData, err := RuleBundleBytes(RuleBundle{FormatVersion: "1", RuleVersion: "test", Rules: []vulnerability.Rule{}})
	if err != nil {
		t.Fatalf("RuleBundleBytes() error = %v", err)
	}
	var ruleObject map[string]interface{}
	if err := json.Unmarshal(ruleData, &ruleObject); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	result, err := ScanInventoryValue(
		map[string]interface{}{"host": "dynamic"},
		[]interface{}{"vuln"},
		map[string]interface{}{"rule_bundle": ruleObject},
	)
	if err != nil {
		t.Fatalf("ScanInventoryValue() error = %v", err)
	}
	if result.Target.Host != "dynamic" || result.RuleSource.RuleVersion != "test" {
		t.Fatalf("dynamic result = %#v", result)
	}
}

func TestScanFilesystemKeepsParserErrorsAndFindsComponents(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"demo":"1.2.3"}}`), 0o600); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "broken.json"), []byte("ignored"), 0o600); err != nil {
		t.Fatalf("write broken.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "requirements.txt"), []byte("requests==2.31.0\n# comment\n"), 0o600); err != nil {
		t.Fatalf("write requirements.txt: %v", err)
	}
	result, err := ScanFilesystem(context.Background(), root, Options{Scanners: []Scanner{ScannerSBOM}})
	if err != nil {
		t.Fatalf("ScanFilesystem() error = %v", err)
	}
	if len(result.Components) != 2 {
		t.Fatalf("components = %#v", result.Components)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("errors = %#v", result.Errors)
	}
}

func TestScanFilesystemPreservesComponentsWhenManifestIsMalformed(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{"), 0o600); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "requirements.txt"), []byte("requests==2.31.0\n"), 0o600); err != nil {
		t.Fatalf("write requirements.txt: %v", err)
	}
	result, err := ScanFilesystem(context.Background(), root, Options{Scanners: []Scanner{ScannerSBOM}})
	if err != nil {
		t.Fatalf("ScanFilesystem() error = %v", err)
	}
	if result.Status != StatusPartial || len(result.Components) != 1 || len(result.Errors) != 1 || result.Errors[0].Code != "parse_error" {
		t.Fatalf("partial result = %#v", result)
	}
}

func TestScanFilesystemEnforcesLimitsAndIgnorePaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "requirements.txt"), []byte("requests==2.31.0\n"), 0o600); err != nil {
		t.Fatalf("write requirements.txt: %v", err)
	}
	ignored := filepath.Join(root, "ignored")
	if err := os.Mkdir(ignored, 0o700); err != nil {
		t.Fatalf("mkdir ignored: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ignored, "package.json"), []byte(`{"dependencies":{"hidden":"1.0.0"}}`), 0o600); err != nil {
		t.Fatalf("write ignored package.json: %v", err)
	}
	result, err := ScanFilesystem(context.Background(), root, Options{Scanners: []Scanner{ScannerSBOM}, MaxFileSize: 4, IgnorePaths: []string{"ignored"}})
	if err != nil {
		t.Fatalf("ScanFilesystem() error = %v", err)
	}
	if result.Status != StatusPartial || len(result.Components) != 0 || len(result.Errors) != 1 || result.Errors[0].Code != "limit_exceeded" {
		t.Fatalf("limited result = %#v", result)
	}
}

func TestScanFilesystemMergesDuplicateComponentsAndLocations(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"app-a", "app-b"} {
		path := filepath.Join(root, directory)
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", directory, err)
		}
		if err := os.WriteFile(filepath.Join(path, "package.json"), []byte(`{"dependencies":{"demo":"1.0.0"}}`), 0o600); err != nil {
			t.Fatalf("write %s package.json: %v", directory, err)
		}
	}
	result, err := ScanFilesystem(context.Background(), root, Options{Scanners: []Scanner{ScannerSBOM}})
	if err != nil {
		t.Fatalf("ScanFilesystem() error = %v", err)
	}
	if len(result.Components) != 1 || len(result.Components[0].Locations) != 2 {
		t.Fatalf("merged components = %#v", result.Components)
	}
}

func TestScanFilesystemHandlesEmptyDirectoryAndCancellation(t *testing.T) {
	root := t.TempDir()
	result, err := ScanFilesystem(context.Background(), root, Options{Scanners: []Scanner{ScannerSBOM}})
	if err != nil {
		t.Fatalf("empty ScanFilesystem() error = %v", err)
	}
	if result.Components == nil || len(result.Components) != 0 {
		t.Fatalf("empty components = %#v", result.Components)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err = ScanFilesystem(ctx, root, Options{Scanners: []Scanner{ScannerSBOM}})
	if err != nil {
		t.Fatalf("cancelled ScanFilesystem() error = %v", err)
	}
	if result.Status != StatusPartial || len(result.Errors) == 0 || result.Errors[0].Code != "cancelled" {
		t.Fatalf("cancelled result = %#v", result)
	}
}

func TestGenerateSBOMFormats(t *testing.T) {
	components := []SBOMComponent{{ID: "go:demo:1.0", Ecosystem: "go", Name: "demo", Version: "1.0"}}
	for _, format := range []string{"json", "cyclonedx", "spdx"} {
		data, err := EncodeSBOM(components, format)
		if err != nil || len(data) == 0 {
			t.Fatalf("EncodeSBOM(%q) data=%q err=%v", format, data, err)
		}
	}
	if _, err := GenerateSBOM(components, "unknown"); err == nil {
		t.Fatal("GenerateSBOM() accepted an unknown format")
	}
	cycloneValue, err := GenerateSBOM(components, "cyclonedx")
	if err != nil {
		t.Fatalf("GenerateSBOM(cyclonedx) error = %v", err)
	}
	cyclone := cycloneValue.(map[string]interface{})
	cycloneComponents := cyclone["components"].([]map[string]interface{})
	if cycloneComponents[0]["type"] != "library" || cycloneComponents[0]["bom-ref"] != "go:demo:1.0" {
		t.Fatalf("CycloneDX component = %#v", cycloneComponents[0])
	}
	spdxValue, err := GenerateSBOM(components, "spdx")
	if err != nil {
		t.Fatalf("GenerateSBOM(spdx) error = %v", err)
	}
	spdx := spdxValue.(map[string]interface{})
	spdxPackages := spdx["packages"].([]map[string]interface{})
	if spdx["documentNamespace"] == "" || spdxPackages[0]["SPDXID"] == "" || spdxPackages[0]["downloadLocation"] != "NOASSERTION" {
		t.Fatalf("SPDX document = %#v", spdx)
	}
}
