package securityscan

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/software"
)

// Capability identifies one class of security scanner behavior.
type Capability string

const (
	CapabilityVulnerability Capability = "vulnerability"
	CapabilitySBOM          Capability = "sbom"
	CapabilityMisconfig     Capability = "misconfig"
	CapabilitySecret        Capability = "secret"
	CapabilityLicense       Capability = "license"
)

// ScannerBackend is the stable contract between OpsLang and scan engines.
// Implementations may execute locally, through a temporary binary, or through
// a controller service.
type ScannerBackend interface {
	Name() string
	Capabilities() []Capability
	Scan(context.Context, ScanRequest) (ScanReport, error)
}

type ScanTarget struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Path string `json:"path,omitempty"`
	Host string `json:"host,omitempty"`
}

type ScanRequest struct {
	Inventory       *software.InventoryResult `json:"inventory,omitempty"`
	Target          ScanTarget                `json:"target"`
	Scanners        []Capability              `json:"scanners"`
	Severity        string                    `json:"severity,omitempty"`
	Format          string                    `json:"format,omitempty"`
	IgnoreRules     []IgnoreRule              `json:"ignore_rules,omitempty"`
	IgnorePaths     []string                  `json:"ignore_paths,omitempty"`
	MaxFileSize     int64                     `json:"max_file_size,omitempty"`
	MaxDepth        int                       `json:"max_depth,omitempty"`
	Timeout         time.Duration             `json:"timeout,omitempty"`
	DatabaseName    string                    `json:"database_name,omitempty"`
	DatabaseVersion string                    `json:"database_version,omitempty"`
}

type Coverage struct {
	Complete bool     `json:"complete"`
	Scanned  int64    `json:"scanned,omitempty"`
	Skipped  int64    `json:"skipped,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

type PolicyResult struct {
	Status       string `json:"status"`
	Blocked      bool   `json:"blocked"`
	Reason       string `json:"reason,omitempty"`
	FindingCount int    `json:"finding_count"`
}

type ScanReport struct {
	DatabaseSHA256  string                 `json:"database_sha256,omitempty"`
	Backend         string                 `json:"backend"`
	BackendVersion  string                 `json:"backend_version,omitempty"`
	DatabaseName    string                 `json:"database_name,omitempty"`
	DatabaseVersion string                 `json:"database_version,omitempty"`
	Target          ScanTarget             `json:"target"`
	Status          string                 `json:"status"`
	Findings        []VulnerabilityFinding `json:"findings,omitempty"`
	Components      []SBOMComponent        `json:"components,omitempty"`
	Errors          []ScanError            `json:"errors,omitempty"`
	Coverage        Coverage               `json:"coverage"`
	Policy          PolicyResult           `json:"policy"`
	StartedAt       time.Time              `json:"started_at"`
	FinishedAt      time.Time              `json:"finished_at"`
}

func ValidateScanRequest(request ScanRequest) error {
	if strings.TrimSpace(request.Target.ID) == "" {
		return fmt.Errorf("scan target id is required")
	}
	if len(request.Scanners) == 0 {
		return fmt.Errorf("at least one scanner capability is required")
	}
	for _, capability := range request.Scanners {
		switch capability {
		case CapabilityVulnerability, CapabilitySBOM, CapabilityMisconfig, CapabilitySecret, CapabilityLicense:
		default:
			return fmt.Errorf("unknown scanner capability %q", capability)
		}
	}
	if request.MaxFileSize < 0 || request.MaxDepth < 0 || request.Timeout < 0 {
		return fmt.Errorf("scan limits must not be negative")
	}
	return nil
}

// ValidateAllowedCapabilities applies the policy configured by the caller to
// a validated scan request before any scanner backend is invoked.
func ValidateAllowedCapabilities(request ScanRequest, allowed []Capability) error {
	if err := ValidateScanRequest(request); err != nil {
		return err
	}
	if len(allowed) == 0 {
		return fmt.Errorf("no scanner capabilities are allowed")
	}
	permitted := make(map[Capability]struct{}, len(allowed))
	for _, capability := range allowed {
		permitted[capability] = struct{}{}
	}
	for _, capability := range request.Scanners {
		if _, ok := permitted[capability]; !ok {
			return fmt.Errorf("scanner capability %q is not allowed", capability)
		}
	}
	return nil
}

func (r *ScanReport) Finish() {
	if r == nil {
		return
	}
	if r.Findings == nil {
		r.Findings = []VulnerabilityFinding{}
	}
	if r.Components == nil {
		r.Components = []SBOMComponent{}
	}
	if r.Errors == nil {
		r.Errors = []ScanError{}
	}
	if r.FinishedAt.IsZero() {
		r.FinishedAt = time.Now().UTC()
	}
}
