package zypper

import (
	"encoding/json"
	"testing"
)

// TestResultJSON tests JSON marshaling of Result.
func TestResultJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  Result
		fields []string
	}{
		{
			name:   "success",
			input:  Result{Status: "success", Changed: true, Package: "vim"},
			fields: []string{"status", "changed", "package"},
		},
		{
			name:   "failed",
			input:  Result{Status: "failed", Error: "not found"},
			fields: []string{"status", "error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			var result Result
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if result.Status != tt.input.Status {
				t.Errorf("Status: got %q, want %q", result.Status, tt.input.Status)
			}
			if result.Changed != tt.input.Changed {
				t.Errorf("Changed: got %v, want %v", result.Changed, tt.input.Changed)
			}
			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("unmarshal raw failed: %v", err)
			}
			for _, f := range tt.fields {
				if _, ok := raw[f]; !ok {
					t.Errorf("missing expected field: %s", f)
				}
			}
		})
	}
}

// TestInfoResultJSON tests JSON marshaling of InfoResult.
func TestInfoResultJSON(t *testing.T) {
	info := InfoResult{
		Name:    "vim",
		Version: "9.0.1",
		Arch:    "x86_64",
		Status:  "installed",
		Summary: "Vi Improved",
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result InfoResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result.Name != "vim" || result.Version != "9.0.1" {
		t.Errorf("unexpected result: %+v", result)
	}
}

// TestRepoInfoJSON tests JSON marshaling of RepoInfo.
func TestRepoInfoJSON(t *testing.T) {
	repo := RepoInfo{
		Name:    "oss",
		URL:     "http://download.opensuse.org/distribution/leap/15.5/repo/oss/",
		Enabled: "Yes",
		Type:    "rpm-md",
	}
	data, err := json.Marshal(repo)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result RepoInfo
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result.Name != "oss" || result.URL != repo.URL {
		t.Errorf("unexpected result: %+v", result)
	}
}

// TestInstallValidation tests that Install validates required arguments.
func TestInstallValidation(t *testing.T) {
	_, err := Install("", "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestRemoveValidation tests that Remove validates required arguments.
func TestRemoveValidation(t *testing.T) {
	_, err := Remove("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestInfoValidation tests that Info validates required arguments.
func TestInfoValidation(t *testing.T) {
	_, err := Info("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestRepoAddValidation tests that RepoAdd validates required arguments.
func TestRepoAddValidation(t *testing.T) {
	_, err := RepoAdd("", "http://example.com")
	if err == nil {
		t.Error("expected error for empty name")
	}
	_, err = RepoAdd("myrepo", "")
	if err == nil {
		t.Error("expected error for empty url")
	}
}

// TestRepoRemoveValidation tests that RepoRemove validates required arguments.
func TestRepoRemoveValidation(t *testing.T) {
	_, err := RepoRemove("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestSearchValidation tests that Search validates required arguments.
func TestSearchValidation(t *testing.T) {
	_, err := Search("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestPatternInstallValidation tests that PatternInstall validates required arguments.
func TestPatternInstallValidation(t *testing.T) {
	_, err := PatternInstall("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestPatternRemoveValidation tests that PatternRemove validates required arguments.
func TestPatternRemoveValidation(t *testing.T) {
	_, err := PatternRemove("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}
