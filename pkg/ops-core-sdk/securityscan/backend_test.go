package securityscan

import "testing"

func TestValidateScanRequest(t *testing.T) {
	tests := []struct {
		name    string
		request ScanRequest
		wantErr bool
	}{
		{name: "valid", request: ScanRequest{Target: ScanTarget{ID: "host-1"}, Scanners: []Capability{CapabilitySBOM}}},
		{name: "missing target", request: ScanRequest{Scanners: []Capability{CapabilitySBOM}}, wantErr: true},
		{name: "missing scanner", request: ScanRequest{Target: ScanTarget{ID: "host-1"}}, wantErr: true},
		{name: "negative limit", request: ScanRequest{Target: ScanTarget{ID: "host-1"}, Scanners: []Capability{CapabilitySBOM}, MaxDepth: -1}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateScanRequest(test.request)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateScanRequest() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateAllowedCapabilities(t *testing.T) {
	request := ScanRequest{Target: ScanTarget{ID: "host"}, Scanners: []Capability{CapabilityVulnerability}}
	if err := ValidateAllowedCapabilities(request, []Capability{CapabilityVulnerability}); err != nil {
		t.Fatalf("allowed capability rejected: %v", err)
	}
	if err := ValidateAllowedCapabilities(request, []Capability{CapabilitySBOM}); err == nil {
		t.Fatal("expected capability policy rejection")
	}
}

func TestScanReportFinishInitializesCollections(t *testing.T) {
	report := ScanReport{}
	report.Finish()
	if report.Findings == nil || report.Components == nil || report.Errors == nil {
		t.Fatal("Finish() must initialize report collections")
	}
	if report.FinishedAt.IsZero() {
		t.Fatal("Finish() must set FinishedAt")
	}
}
