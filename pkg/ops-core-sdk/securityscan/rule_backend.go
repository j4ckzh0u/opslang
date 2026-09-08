package securityscan

import (
	"context"
	"fmt"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
)

// RuleBundleBackend provides the built-in database through the common backend API.
type RuleBundleBackend struct {
	Bundle RuleBundle
}

func NewRuleBundleBackend(bundle []byte) (*RuleBundleBackend, error) {
	validated, _, err := ValidateRuleBundle(bundle)
	if err != nil {
		return nil, fmt.Errorf("validate scanner rule bundle: %w", err)
	}
	return &RuleBundleBackend{Bundle: validated}, nil
}

func (b *RuleBundleBackend) Name() string { return "native-rule-bundle" }

func (b *RuleBundleBackend) Capabilities() []Capability {
	return []Capability{CapabilityVulnerability, CapabilitySBOM}
}

func (b *RuleBundleBackend) Scan(ctx context.Context, request ScanRequest) (ScanReport, error) {
	if b == nil {
		return ScanReport{}, fmt.Errorf("rule bundle backend is nil")
	}
	if err := ValidateScanRequest(request); err != nil {
		return ScanReport{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return ScanReport{}, ctx.Err()
	default:
	}
	for _, capability := range request.Scanners {
		if capability != CapabilityVulnerability && capability != CapabilitySBOM {
			return ScanReport{}, fmt.Errorf("native rule bundle does not support %q", capability)
		}
	}
	report := ScanReport{Backend: b.Name(), BackendVersion: "1", DatabaseVersion: b.Bundle.RuleVersion,
		DatabaseSHA256: b.Bundle.SHA256, Target: request.Target, Status: StatusOK,
		Coverage: Coverage{Complete: true}}
	if request.Inventory == nil {
		return ScanReport{}, fmt.Errorf("native rule bundle requires inventory")
	}
	if containsCapability(request.Scanners, CapabilitySBOM) {
		report.Components = ComponentsFromInventory(*request.Inventory)
	}
	if containsCapability(request.Scanners, CapabilityVulnerability) {
		for _, finding := range vulnerability.Match(*request.Inventory, b.Bundle.Rules) {
			report.Findings = append(report.Findings, VulnerabilityFinding{ID: finding.ID, Package: finding.Package, InstalledVersion: finding.InstalledVersion, FixedVersion: finding.FixedVersion, Severity: finding.Severity, Title: finding.Summary, Status: "affected"})
		}
	}
	report.Finish()
	return report, nil
}

func containsCapability(capabilities []Capability, wanted Capability) bool {
	for _, capability := range capabilities {
		if capability == wanted {
			return true
		}
	}
	return false
}
