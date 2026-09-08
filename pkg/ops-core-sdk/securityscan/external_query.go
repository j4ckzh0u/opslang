package securityscan

import (
	"context"
	"fmt"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/software"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
)

func queryExternalVulnerabilities(ctx context.Context, inventory software.InventoryResult, config *RemoteConfig) ([]vulnerability.Finding, RuleSourceInfo, error) {
	backend := &ServiceBackend{Config: &ServiceConfig{
		URL: config.URL, Token: config.Token, CA: config.CA, Timeout: config.Timeout,
		DatabaseVersion: config.RuleVersion, DatabaseSHA256: config.RuleSHA256,
	}}
	target := ScanTarget{Type: "host", ID: inventory.Host, Host: inventory.Host}
	if target.ID == "" {
		target.ID = "local"
	}
	report, err := backend.Scan(ctx, ScanRequest{Target: target, Inventory: &inventory,
		Scanners: []Capability{CapabilityVulnerability}, DatabaseVersion: config.RuleVersion})
	if err != nil {
		return nil, RuleSourceInfo{}, err
	}
	// The legacy result cannot represent coverage; require a complete response.
	if report.Status != StatusOK || !report.Coverage.Complete || len(report.Errors) > 0 || len(report.Coverage.Errors) > 0 || report.Policy.Blocked {
		return nil, RuleSourceInfo{}, fmt.Errorf("external vulnerability query did not complete successfully (status %q)", report.Status)
	}
	findings := make([]vulnerability.Finding, 0, len(report.Findings))
	for _, finding := range report.Findings {
		findings = append(findings, vulnerability.Finding{ID: finding.ID, Package: finding.Package,
			InstalledVersion: finding.InstalledVersion, FixedVersion: finding.FixedVersion,
			Severity: finding.Severity, Summary: finding.Title})
	}
	return findings, RuleSourceInfo{RuleVersion: report.DatabaseVersion, SHA256: report.DatabaseSHA256}, nil
}
