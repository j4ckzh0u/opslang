package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/securityscan"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/software"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
	"github.com/spf13/cobra"
)

func TestParseScanScanners(t *testing.T) {
	scanners, err := parseScanScanners("vuln, sbom")
	if err != nil {
		t.Fatalf("parseScanScanners() error = %v", err)
	}
	if len(scanners) != 2 || scanners[0] != securityscan.ScannerVulnerability || scanners[1] != securityscan.ScannerSBOM {
		t.Fatalf("scanners = %#v", scanners)
	}
	if _, err := parseScanScanners(""); err == nil {
		t.Fatal("parseScanScanners() accepted an empty list")
	}
	if _, err := parseScanScanners("secret"); err == nil {
		t.Fatal("parseScanScanners() accepted an unsupported scanner")
	}
}

func TestRunScanCommandReturnsGateErrorAfterJSONOutput(t *testing.T) {
	directory := t.TempDir()
	inventoryPath := filepath.Join(directory, "inventory.json")
	rulesPath := filepath.Join(directory, "rules.json")
	inventory, err := json.Marshal(software.InventoryResult{Host: "test", Packages: []software.Package{{Name: "demo", Version: "1.0.0"}}})
	if err != nil {
		t.Fatalf("encode inventory: %v", err)
	}
	if err := os.WriteFile(inventoryPath, inventory, 0o600); err != nil {
		t.Fatalf("write inventory: %v", err)
	}
	rules, err := securityscan.RuleBundleBytes(securityscan.RuleBundle{FormatVersion: "1", RuleVersion: "test", Rules: []vulnerability.Rule{{ID: "CVE-TEST", Package: "demo", FixedVersion: "2.0.0", Severity: "high"}}})
	if err != nil {
		t.Fatalf("encode rules: %v", err)
	}
	if err := os.WriteFile(rulesPath, rules, 0o600); err != nil {
		t.Fatalf("write rules: %v", err)
	}

	originalPath, originalInventory, originalRules := scanPath, scanInventory, scanRules
	originalScanners, originalFormat, originalSeverity := scanScanners, scanFormat, scanSeverity
	originalMaxSize, originalMaxDepth := scanMaxSize, scanMaxDepth
	originalIgnoreIDs, originalIgnorePkgs, originalIgnorePaths := scanIgnoreIDs, scanIgnorePkgs, scanIgnorePaths
	t.Cleanup(func() {
		scanPath, scanInventory, scanRules = originalPath, originalInventory, originalRules
		scanScanners, scanFormat, scanSeverity = originalScanners, originalFormat, originalSeverity
		scanMaxSize, scanMaxDepth = originalMaxSize, originalMaxDepth
		scanIgnoreIDs, scanIgnorePkgs, scanIgnorePaths = originalIgnoreIDs, originalIgnorePkgs, originalIgnorePaths
	})
	scanPath, scanInventory, scanRules = "", inventoryPath, rulesPath
	scanScanners, scanFormat, scanSeverity = "vuln", "json", "high"
	scanMaxSize, scanMaxDepth = 0, 0
	scanIgnoreIDs, scanIgnorePkgs, scanIgnorePaths = nil, nil, nil

	var output bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&output)
	err = runScanCommand(command)
	if !errors.Is(err, errScanGate) {
		t.Fatalf("runScanCommand() error = %v", err)
	}
	if !bytes.Contains(output.Bytes(), []byte(`"status": "failed"`)) || !bytes.Contains(output.Bytes(), []byte(`"id": "CVE-TEST"`)) {
		t.Fatalf("scan output = %s", output.String())
	}
}
