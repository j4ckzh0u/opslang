package securityscan

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/software"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
)

func ScanInventoryValue(inventoryValue, scannersValue, optionsValue interface{}) (ScanResult, error) {
	inventory, err := inventoryFromValue(inventoryValue)
	if err != nil {
		return ScanResult{}, err
	}
	options, err := optionsFromValue(scannersValue, optionsValue)
	if err != nil {
		return ScanResult{}, err
	}
	return ScanInventory(inventory, options)
}

func ScanFilesystemValue(path string, scannersValue, optionsValue interface{}) (ScanResult, error) {
	options, err := optionsFromValue(scannersValue, optionsValue)
	if err != nil {
		return ScanResult{}, err
	}
	return ScanFilesystem(context.Background(), path, options)
}

func SBOMValue(inventoryValue, formatValue interface{}) (interface{}, error) {
	inventory, err := inventoryFromValue(inventoryValue)
	if err != nil {
		return nil, err
	}
	format, ok := formatValue.(string)
	if !ok {
		return nil, fmt.Errorf("SBOM format must be a string, got %T", formatValue)
	}
	return GenerateSBOM(ComponentsFromInventory(inventory), format)
}

func inventoryFromValue(value interface{}) (software.InventoryResult, error) {
	switch typed := value.(type) {
	case software.InventoryResult:
		return typed, nil
	case *software.InventoryResult:
		if typed == nil {
			return software.InventoryResult{}, fmt.Errorf("inventory must not be nil")
		}
		return *typed, nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return software.InventoryResult{}, fmt.Errorf("encode inventory: %w", err)
		}
		var inventory software.InventoryResult
		if err := json.Unmarshal(data, &inventory); err != nil {
			return software.InventoryResult{}, fmt.Errorf("decode inventory: %w", err)
		}
		return inventory, nil
	}
}

func optionsFromValue(scannersValue, optionsValue interface{}) (Options, error) {
	options := Options{Scanners: make([]Scanner, 0)}
	if scanners, ok := scannersValue.([]Scanner); ok {
		options.Scanners = append(options.Scanners, scanners...)
	} else {
		data, err := json.Marshal(scannersValue)
		if err != nil {
			return Options{}, fmt.Errorf("encode scanners: %w", err)
		}
		var values []string
		if err := json.Unmarshal(data, &values); err != nil {
			return Options{}, fmt.Errorf("scanners must be a string list: %w", err)
		}
		for _, value := range values {
			options.Scanners = append(options.Scanners, Scanner(value))
		}
	}
	if optionsValue != nil {
		var extra Options
		switch typed := optionsValue.(type) {
		case Options:
			extra = typed
		case *Options:
			if typed == nil {
				return Options{}, fmt.Errorf("scan options must not be nil")
			}
			extra = *typed
		default:
			data, err := json.Marshal(optionsValue)
			if err != nil {
				return Options{}, fmt.Errorf("encode scan options: %w", err)
			}
			if err := json.Unmarshal(data, &extra); err != nil {
				return Options{}, fmt.Errorf("decode scan options: %w", err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil {
				return Options{}, fmt.Errorf("decode scan option fields: %w", err)
			}
			if raw, ok := fields["rule_bundle"]; ok {
				extra.RuleBundle, err = ruleBundleFromJSON(raw)
				if err != nil {
					return Options{}, err
				}
			}
		}
		extra.Scanners = options.Scanners
		options = extra
	}
	return options, nil
}

func ruleBundleFromJSON(raw json.RawMessage) ([]byte, error) {
	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "{") {
		return append([]byte(nil), raw...), nil
	}
	var encoded []byte
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil, fmt.Errorf("scan option rule_bundle must be an object or base64 bytes: %w", err)
	}
	return encoded, nil
}

func ScanInventory(inventory software.InventoryResult, options Options) (ScanResult, error) {
	if err := ValidateOptions(options); err != nil {
		return ScanResult{}, err
	}
	result := NewScanResult(TargetInfo{Type: "host", ID: inventory.Host, Host: inventory.Host}, options.Scanners)
	if result.Target.ID == "" {
		result.Target.ID = "local"
	}
	if containsScanner(options.Scanners, ScannerVulnerability) {
		var err error
		var findings []vulnerability.Finding
		var source RuleSourceInfo
		if options.Remote != nil {
			findings, source, err = queryRemoteVulnerabilities(context.Background(), inventory, options.Remote)
		} else {
			bundle, localSource, localErr := ValidateRuleBundle(options.RuleBundle)
			err = localErr
			if err == nil {
				source = localSource
				findings = vulnerability.Match(inventory, bundle.Rules)
			}
		}
		if err != nil {
			return ScanResult{}, fmt.Errorf("scan inventory: %w", err)
		}
		result.RuleSource = source
		for _, finding := range findings {
			if !severityAllowed(finding.Severity, options.Severity) {
				continue
			}
			converted := VulnerabilityFinding{
				ID:               finding.ID,
				Package:          finding.Package,
				InstalledVersion: finding.InstalledVersion,
				FixedVersion:     finding.FixedVersion,
				Severity:         finding.Severity,
				Title:            finding.Summary,
				Path:             inventoryPackagePath(inventory, finding.Package, finding.InstalledVersion),
				Status:           "affected",
			}
			if !isIgnoredFinding(converted, options.IgnoreRules) {
				result.Findings = append(result.Findings, converted)
			}
		}
	}
	if containsScanner(options.Scanners, ScannerSBOM) {
		result.Components = ComponentsFromInventory(inventory)
	}
	StableSortFindings(result.Findings)
	StableSortComponents(result.Components)
	applySeverityGate(&result, options.Severity)
	result.Finish()
	return result, nil
}

func ScanFilesystem(ctx context.Context, path string, options Options) (ScanResult, error) {
	if err := ValidateOptions(options); err != nil {
		return ScanResult{}, err
	}
	if strings.TrimSpace(path) == "" {
		return ScanResult{}, fmt.Errorf("scan target path is required")
	}
	info, err := os.Stat(path)
	if err != nil {
		return ScanResult{}, fmt.Errorf("stat scan target: %w", err)
	}
	if !info.IsDir() {
		return ScanResult{}, fmt.Errorf("scan target path must be a directory")
	}
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}
	var source RuleSourceInfo
	result := NewScanResult(TargetInfo{Type: "filesystem", ID: filepath.Clean(path), Path: filepath.Clean(path)}, options.Scanners)
	components, scanErrors := scanManifests(ctx, path, options)
	result.Errors = append(result.Errors, scanErrors...)
	if containsScanner(options.Scanners, ScannerSBOM) || containsScanner(options.Scanners, ScannerVulnerability) {
		result.Components = components
	}
	if containsScanner(options.Scanners, ScannerVulnerability) {
		inventory := software.InventoryResult{Host: filepath.Clean(path), Packages: make([]software.Package, 0, len(components))}
		for _, component := range components {
			inventory.Packages = append(inventory.Packages, software.Package{Name: component.Name, Version: component.Version, Manager: component.Ecosystem, InstalledFiles: append([]string(nil), component.Locations...)})
		}
		var findings []vulnerability.Finding
		if options.Remote != nil {
			findings, source, err = queryRemoteVulnerabilities(ctx, inventory, options.Remote)
			if err != nil {
				return ScanResult{}, fmt.Errorf("scan filesystem: %w", err)
			}
		} else {
			bundle, bundleSource, bundleErr := ValidateRuleBundle(options.RuleBundle)
			if bundleErr != nil {
				return ScanResult{}, fmt.Errorf("scan filesystem: %w", bundleErr)
			}
			source = bundleSource
			findings = vulnerability.Match(inventory, bundle.Rules)
		}
		result.RuleSource = source
		for _, finding := range findings {
			if !severityAllowed(finding.Severity, options.Severity) {
				continue
			}
			converted := VulnerabilityFinding{ID: finding.ID, Package: finding.Package, InstalledVersion: finding.InstalledVersion, FixedVersion: finding.FixedVersion, Severity: finding.Severity, Title: finding.Summary, Path: inventoryPackagePath(inventory, finding.Package, finding.InstalledVersion), Status: "affected"}
			if !isIgnoredFinding(converted, options.IgnoreRules) {
				result.Findings = append(result.Findings, converted)
			}
		}
	}
	StableSortFindings(result.Findings)
	StableSortComponents(result.Components)
	if len(result.Errors) > 0 {
		result.Status = StatusPartial
	}
	applySeverityGate(&result, options.Severity)
	result.Finish()
	return result, nil
}

func ComponentsFromInventory(inventory software.InventoryResult) []SBOMComponent {
	byID := make(map[string]SBOMComponent)
	for _, item := range inventory.Packages {
		ecosystem := strings.ToLower(strings.TrimSpace(item.Manager))
		if ecosystem == "" {
			ecosystem = "system"
		}
		id := ComponentID(ecosystem, item.Name, item.Version)
		component := byID[id]
		if component.ID == "" {
			component = SBOMComponent{ID: id, Name: item.Name, Version: item.Version, Ecosystem: ecosystem, Source: "software.inventory"}
		}
		component.Locations = appendUnique(component.Locations, item.InstalledFiles...)
		byID[id] = component
	}
	components := make([]SBOMComponent, 0, len(byID))
	for _, component := range byID {
		components = append(components, component)
	}
	StableSortComponents(components)
	return components
}

func GenerateSBOM(components []SBOMComponent, format string) (interface{}, error) {
	components = append([]SBOMComponent(nil), components...)
	StableSortComponents(components)
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "native", "json":
		return components, nil
	case "cyclonedx":
		return cycloneDXDocument(components), nil
	case "spdx":
		return spdxDocument(components)
	default:
		return nil, fmt.Errorf("unsupported SBOM format %q", format)
	}
}

func cycloneDXDocument(components []SBOMComponent) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(components))
	for _, component := range components {
		item := map[string]interface{}{
			"type":    "library",
			"bom-ref": component.ID,
			"name":    component.Name,
			"version": component.Version,
		}
		properties := []map[string]string{{"name": "opslang:ecosystem", "value": component.Ecosystem}}
		if component.Source != "" {
			properties = append(properties, map[string]string{"name": "opslang:source", "value": component.Source})
		}
		for _, location := range component.Locations {
			properties = append(properties, map[string]string{"name": "opslang:location", "value": location})
		}
		item["properties"] = properties
		if len(component.Licenses) > 0 {
			licenses := make([]map[string]interface{}, 0, len(component.Licenses))
			for _, license := range component.Licenses {
				licenses = append(licenses, map[string]interface{}{"license": map[string]string{"name": license}})
			}
			item["licenses"] = licenses
		}
		items = append(items, item)
	}
	return map[string]interface{}{"bomFormat": "CycloneDX", "specVersion": "1.5", "version": 1, "components": items}
}

func spdxDocument(components []SBOMComponent) (map[string]interface{}, error) {
	packages := make([]map[string]interface{}, 0, len(components))
	for _, component := range components {
		item := map[string]interface{}{
			"name":             component.Name,
			"SPDXID":           spdxPackageID(component.ID),
			"versionInfo":      component.Version,
			"downloadLocation": "NOASSERTION",
			"filesAnalyzed":    false,
		}
		if len(component.Licenses) > 0 {
			item["licenseConcluded"] = "NOASSERTION"
			item["licenseDeclared"] = "NOASSERTION"
		}
		packages = append(packages, item)
	}
	encoded, err := json.Marshal(components)
	if err != nil {
		return nil, fmt.Errorf("encode components for SPDX namespace: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return map[string]interface{}{
		"spdxVersion":       "SPDX-2.3",
		"dataLicense":       "CC0-1.0",
		"SPDXID":            "SPDXRef-DOCUMENT",
		"name":              "OpsLang SBOM",
		"documentNamespace": fmt.Sprintf("https://opslang.dev/sbom/%x", digest),
		"creationInfo": map[string]interface{}{
			"created":  "1970-01-01T00:00:00Z",
			"creators": []string{"Tool: OpsLang"},
		},
		"packages": packages,
	}, nil
}

func spdxPackageID(componentID string) string {
	digest := sha256.Sum256([]byte(componentID))
	return fmt.Sprintf("SPDXRef-Package-%x", digest[:8])
}

func EncodeSBOM(components []SBOMComponent, format string) ([]byte, error) {
	value, err := GenerateSBOM(components, format)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode SBOM: %w", err)
	}
	return data, nil
}

func containsScanner(scanners []Scanner, wanted Scanner) bool {
	for _, scanner := range scanners {
		if scanner == wanted {
			return true
		}
	}
	return false
}

func severityAllowed(severity, minimum string) bool {
	if strings.TrimSpace(minimum) == "" {
		return true
	}
	ranks := map[string]int{"unknown": 0, "negligible": 1, "low": 2, "medium": 3, "high": 4, "critical": 5}
	return ranks[strings.ToLower(strings.TrimSpace(severity))] >= ranks[strings.ToLower(strings.TrimSpace(minimum))]
}

func applySeverityGate(result *ScanResult, minimum string) {
	if result == nil || strings.TrimSpace(minimum) == "" || len(result.Findings) == 0 {
		return
	}
	result.Status = StatusFailed
}

func isIgnoredFinding(finding VulnerabilityFinding, rules []IgnoreRule) bool {
	for _, rule := range rules {
		if rule.ID == "" && rule.Package == "" && rule.Path == "" {
			continue
		}
		if rule.ID != "" && !strings.EqualFold(strings.TrimSpace(rule.ID), finding.ID) {
			continue
		}
		if rule.Package != "" && !strings.EqualFold(strings.TrimSpace(rule.Package), finding.Package) {
			continue
		}
		if rule.Path != "" && filepath.Clean(strings.TrimSpace(rule.Path)) != filepath.Clean(finding.Path) {
			continue
		}
		return true
	}
	return false
}

func inventoryPackagePath(inventory software.InventoryResult, name, version string) string {
	for _, item := range inventory.Packages {
		if item.Name == name && item.Version == version && len(item.InstalledFiles) > 0 {
			return item.InstalledFiles[0]
		}
	}
	return ""
}

func appendUnique(values []string, additions ...string) []string {
	seen := make(map[string]struct{}, len(values)+len(additions))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	for _, value := range additions {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
