package pacman

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
		{
			name:   "no change",
			input:  Result{Status: "success", Changed: false, Output: "no orphans"},
			fields: []string{"status", "changed", "output"},
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
		Size:    "3.5 MiB",
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
	if result.Name != "vim" || result.Version != "9.0.1" || result.Size != "3.5 MiB" {
		t.Errorf("unexpected result: %+v", result)
	}
}

// TestInstallValidation tests that Install validates required arguments.
func TestInstallValidation(t *testing.T) {
	_, err := Install("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestRemoveValidation tests that Remove validates required arguments.
func TestRemoveValidation(t *testing.T) {
	_, err := Remove("", false)
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

// TestSearchValidation tests that Search validates required arguments.
func TestSearchValidation(t *testing.T) {
	_, err := Search("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestInstallFileValidation tests that InstallFile validates required arguments.
func TestInstallFileValidation(t *testing.T) {
	_, err := InstallFile("")
	if err == nil {
		t.Error("expected error for empty path")
	}
}
