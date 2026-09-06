// Package securityscan defines shared models and validation for local scans.
package securityscan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
)

type Scanner string

const (
	ScannerVulnerability Scanner = "vuln"
	ScannerSBOM          Scanner = "sbom"
	ScannerMisconfig     Scanner = "misconfig"
	ScannerSecret        Scanner = "secret"
	ScannerLicense       Scanner = "license"
)

const (
	StatusOK          = "ok"
	StatusPartial     = "partial"
	StatusUnsupported = "unsupported"
	StatusFailed      = "failed"
)

type IgnoreRule struct {
	ID      string `json:"id,omitempty"`
	Package string `json:"package,omitempty"`
	Path    string `json:"path,omitempty"`
}

type Options struct {
	Scanners       []Scanner     `json:"scanners"`
	RuleBundle     []byte        `json:"-"`
	RuleBundleHash string        `json:"rule_bundle_hash,omitempty"`
	IgnoreRules    []IgnoreRule  `json:"ignore_rules,omitempty"`
	IgnorePaths    []string      `json:"ignore_paths,omitempty"`
	MaxFileSize    int64         `json:"max_file_size,omitempty"`
	MaxDepth       int           `json:"max_depth,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty"`
	Severity       string        `json:"severity,omitempty"`
	Format         string        `json:"format,omitempty"`
}

type TargetInfo struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Path string `json:"path,omitempty"`
	Host string `json:"host,omitempty"`
}

type RuleSourceInfo struct {
	FormatVersion string `json:"format_version"`
	RuleVersion   string `json:"rule_version"`
	SHA256        string `json:"sha256"`
}

type VulnerabilityFinding struct {
	ID               string `json:"id"`
	Package          string `json:"package"`
	InstalledVersion string `json:"installed_version"`
	FixedVersion     string `json:"fixed_version,omitempty"`
	Severity         string `json:"severity,omitempty"`
	Title            string `json:"title,omitempty"`
	Path             string `json:"path,omitempty"`
	Status           string `json:"status"`
}

type SBOMComponent struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Ecosystem string   `json:"ecosystem"`
	Locations []string `json:"locations,omitempty"`
	Licenses  []string `json:"licenses,omitempty"`
	Source    string   `json:"source,omitempty"`
}

type ScanError struct {
	Path    string `json:"path,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ScanResult struct {
	Target     TargetInfo             `json:"target"`
	Scanners   []Scanner              `json:"scanners"`
	Status     string                 `json:"status"`
	Findings   []VulnerabilityFinding `json:"findings,omitempty"`
	Components []SBOMComponent        `json:"components,omitempty"`
	Errors     []ScanError            `json:"errors,omitempty"`
	RuleSource RuleSourceInfo         `json:"rule_source,omitempty"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
}

type RuleBundle struct {
	FormatVersion string               `json:"format_version"`
	RuleVersion   string               `json:"rule_version"`
	GeneratedAt   time.Time            `json:"generated_at"`
	Rules         []vulnerability.Rule `json:"rules"`
	SHA256        string               `json:"sha256"`
}

func NewScanResult(target TargetInfo, scanners []Scanner) ScanResult {
	return ScanResult{
		Target:     target,
		Scanners:   append([]Scanner(nil), scanners...),
		Status:     StatusOK,
		Findings:   make([]VulnerabilityFinding, 0),
		Components: make([]SBOMComponent, 0),
		Errors:     make([]ScanError, 0),
		StartedAt:  time.Now().UTC(),
	}
}

func (r *ScanResult) Finish() {
	if r == nil {
		return
	}
	if r.Findings == nil {
		r.Findings = make([]VulnerabilityFinding, 0)
	}
	if r.Components == nil {
		r.Components = make([]SBOMComponent, 0)
	}
	if r.Errors == nil {
		r.Errors = make([]ScanError, 0)
	}
	if r.FinishedAt.IsZero() {
		r.FinishedAt = time.Now().UTC()
	}
}

func ValidateScanners(scanners []Scanner) error {
	if len(scanners) == 0 {
		return fmt.Errorf("at least one scanner is required")
	}
	seen := make(map[Scanner]struct{}, len(scanners))
	for _, scanner := range scanners {
		if _, ok := seen[scanner]; ok {
			continue
		}
		seen[scanner] = struct{}{}
		switch scanner {
		case ScannerVulnerability, ScannerSBOM:
		case ScannerMisconfig, ScannerSecret, ScannerLicense:
			return fmt.Errorf("unsupported scanner %q", scanner)
		default:
			return fmt.Errorf("unsupported scanner %q", scanner)
		}
	}
	return nil
}

func ValidateOptions(options Options) error {
	if err := ValidateScanners(options.Scanners); err != nil {
		return err
	}
	if options.MaxFileSize < 0 {
		return fmt.Errorf("max_file_size must not be negative")
	}
	if options.MaxDepth < 0 {
		return fmt.Errorf("max_depth must not be negative")
	}
	if options.Timeout < 0 {
		return fmt.Errorf("timeout must not be negative")
	}
	if severity := strings.ToLower(strings.TrimSpace(options.Severity)); severity != "" {
		switch severity {
		case "unknown", "negligible", "low", "medium", "high", "critical":
		default:
			return fmt.Errorf("unsupported severity %q", options.Severity)
		}
	}
	if format := strings.ToLower(strings.TrimSpace(options.Format)); format != "" {
		switch format {
		case "native", "json", "cyclonedx", "spdx":
		default:
			return fmt.Errorf("unsupported scan format %q", options.Format)
		}
	}
	return nil
}

func ValidateRuleBundle(data []byte) (RuleBundle, RuleSourceInfo, error) {
	if len(data) == 0 {
		return RuleBundle{}, RuleSourceInfo{}, fmt.Errorf("rule bundle is empty")
	}
	var bundle RuleBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return RuleBundle{}, RuleSourceInfo{}, fmt.Errorf("decode rule bundle: %w", err)
	}
	if bundle.FormatVersion != "1" {
		return RuleBundle{}, RuleSourceInfo{}, fmt.Errorf("unsupported rule bundle format version %q", bundle.FormatVersion)
	}
	if strings.TrimSpace(bundle.RuleVersion) == "" {
		return RuleBundle{}, RuleSourceInfo{}, fmt.Errorf("rule bundle rule_version is required")
	}
	for index, rule := range bundle.Rules {
		if strings.TrimSpace(rule.ID) == "" || strings.TrimSpace(rule.Package) == "" {
			return RuleBundle{}, RuleSourceInfo{}, fmt.Errorf("rule bundle rules[%d] requires id and package", index)
		}
	}
	actualHash, err := ruleBundleHash(bundle)
	if err != nil {
		return RuleBundle{}, RuleSourceInfo{}, fmt.Errorf("hash rule bundle: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(bundle.SHA256), actualHash) {
		return RuleBundle{}, RuleSourceInfo{}, fmt.Errorf("rule bundle sha256 mismatch: declared %q, calculated %q", bundle.SHA256, actualHash)
	}
	return bundle, RuleSourceInfo{FormatVersion: bundle.FormatVersion, RuleVersion: bundle.RuleVersion, SHA256: actualHash}, nil
}

func RuleBundleBytes(bundle RuleBundle) ([]byte, error) {
	bundle.SHA256 = ""
	data, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("encode rule bundle: %w", err)
	}
	hash, err := ruleBundleHash(bundle)
	if err != nil {
		return nil, fmt.Errorf("hash rule bundle: %w", err)
	}
	bundle.SHA256 = hash
	data, err = json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("encode rule bundle with sha256: %w", err)
	}
	return data, nil
}

func ruleBundleHash(bundle RuleBundle) (string, error) {
	bundle.SHA256 = ""
	data, err := json.Marshal(bundle)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func StableSortFindings(findings []VulnerabilityFinding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].ID != findings[j].ID {
			return findings[i].ID < findings[j].ID
		}
		if findings[i].Package != findings[j].Package {
			return findings[i].Package < findings[j].Package
		}
		if findings[i].InstalledVersion != findings[j].InstalledVersion {
			return findings[i].InstalledVersion < findings[j].InstalledVersion
		}
		return findings[i].Path < findings[j].Path
	})
}

func StableSortComponents(components []SBOMComponent) {
	sort.SliceStable(components, func(i, j int) bool {
		if components[i].Ecosystem != components[j].Ecosystem {
			return components[i].Ecosystem < components[j].Ecosystem
		}
		if components[i].Name != components[j].Name {
			return components[i].Name < components[j].Name
		}
		return components[i].Version < components[j].Version
	})
}

func ComponentID(ecosystem, name, version string) string {
	return strings.Join([]string{strings.ToLower(strings.TrimSpace(ecosystem)), strings.TrimSpace(name), strings.TrimSpace(version)}, ":")
}
