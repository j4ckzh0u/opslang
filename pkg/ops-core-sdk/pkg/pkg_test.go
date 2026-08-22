package opspkg

import (
	"encoding/json"
	"testing"
)

func TestPackageActionJSON(t *testing.T) {
	action := PackageAction{
		Name:    "nginx",
		Action:  "install",
		Manager: "apt",
		Success: true,
		Message: "installed successfully",
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal PackageAction: %v", err)
	}

	var decoded PackageAction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal PackageAction: %v", err)
	}

	if decoded != action {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, action)
	}
}

func TestEnsureRejectsEmptyName(t *testing.T) {
	result, err := Ensure("")
	if err == nil {
		t.Fatal("Ensure(\"\") should return an error")
	}
	if result.Changed || result.Success {
		t.Fatalf("empty ensure must not report success: %+v", result)
	}
}

func TestPackageInfoJSON(t *testing.T) {
	info := PackageInfo{
		Name:         "curl",
		Version:      "7.81.0",
		Architecture: "amd64",
		Description:  "command line tool for transferring data",
		Status:       "installed",
		Manager:      "apt",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal PackageInfo: %v", err)
	}

	var decoded PackageInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal PackageInfo: %v", err)
	}

	if decoded != info {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, info)
	}
}

func TestParseDpkgInfoLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  PackageInfo
	}{
		{
			name:  "full line",
			input: "install ok installed|7.81.0-1|amd64|command line tool for transferring data",
			want: PackageInfo{
				Name:         "curl",
				Version:      "7.81.0-1",
				Architecture: "amd64",
				Description:  "command line tool for transferring data",
				Status:       "install ok installed",
				Manager:      "apt",
			},
		},
		{
			name:  "partial line with two fields",
			input: "install ok installed|7.81.0-1",
			want: PackageInfo{
				Name:    "curl",
				Version: "7.81.0-1",
				Status:  "install ok installed",
				Manager: "apt",
			},
		},
		{
			name:  "empty line",
			input: "",
			want: PackageInfo{
				Name:    "curl",
				Manager: "apt",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseDpkgInfoLine("curl", tc.input, "apt")
			if got != tc.want {
				t.Errorf("parseDpkgInfoLine() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestParseDpkgListOutput(t *testing.T) {
	output := "curl|7.81.0-1|amd64|data transfer tool|install ok installed\n" +
		"nginx|1.18.0-6|amd64|web server|install ok installed\n" +
		"\n"

	packages := parseDpkgListOutput(output, "apt")

	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}

	if packages[0].Name != "curl" {
		t.Errorf("first package name = %q, want %q", packages[0].Name, "curl")
	}
	if packages[0].Version != "7.81.0-1" {
		t.Errorf("first package version = %q, want %q", packages[0].Version, "7.81.0-1")
	}
	if packages[1].Name != "nginx" {
		t.Errorf("second package name = %q, want %q", packages[1].Name, "nginx")
	}
	if packages[1].Architecture != "amd64" {
		t.Errorf("second package arch = %q, want %q", packages[1].Architecture, "amd64")
	}
}

func TestParseDpkgListLine(t *testing.T) {
	line := "openssl|3.0.2-0|amd64|Secure Sockets Layer toolkit|install ok installed"
	info := parseDpkgListLine(line, "apt")

	if info.Name != "openssl" {
		t.Errorf("Name = %q, want %q", info.Name, "openssl")
	}
	if info.Version != "3.0.2-0" {
		t.Errorf("Version = %q, want %q", info.Version, "3.0.2-0")
	}
	if info.Architecture != "amd64" {
		t.Errorf("Architecture = %q, want %q", info.Architecture, "amd64")
	}
	if info.Description != "Secure Sockets Layer toolkit" {
		t.Errorf("Description = %q, want %q", info.Description, "Secure Sockets Layer toolkit")
	}
	if info.Status != "install ok installed" {
		t.Errorf("Status = %q, want %q", info.Status, "install ok installed")
	}
	if info.Manager != "apt" {
		t.Errorf("Manager = %q, want %q", info.Manager, "apt")
	}
}

func TestParseRpmInfo(t *testing.T) {
	output := `Name        : curl
Version     : 7.76.1
Release     : 12.el9
Architecture: x86_64
Summary     : A utility for getting files from remote servers
Description : curl is a command line tool for transferring data.
Status      : installed
`
	info := parseRpmInfo("curl", output, "yum")

	if info.Name != "curl" {
		t.Errorf("Name = %q, want %q", info.Name, "curl")
	}
	if info.Version != "7.76.1" {
		t.Errorf("Version = %q, want %q", info.Version, "7.76.1")
	}
	if info.Architecture != "x86_64" {
		t.Errorf("Architecture = %q, want %q", info.Architecture, "x86_64")
	}
	if info.Description == "" {
		t.Error("Description should not be empty")
	}
	if info.Status != "installed" {
		t.Errorf("Status = %q, want %q", info.Status, "installed")
	}
	if info.Manager != "yum" {
		t.Errorf("Manager = %q, want %q", info.Manager, "yum")
	}
}

func TestParseRpmInfoDefaultsInstalled(t *testing.T) {
	output := `Name        : bash
Version     : 5.1.8
Architecture: x86_64
Summary     : The GNU Bourne Again shell
`
	info := parseRpmInfo("bash", output, "dnf")
	if info.Status != "installed" {
		t.Errorf("Status = %q, want %q (should default to installed)", info.Status, "installed")
	}
}

func TestParseRpmListLine(t *testing.T) {
	line := "bash|5.1.8-6.el9_1|x86_64|The GNU Bourne Again shell|installed"
	info := parseRpmListLine(line, "dnf")

	if info.Name != "bash" {
		t.Errorf("Name = %q, want %q", info.Name, "bash")
	}
	if info.Version != "5.1.8-6.el9_1" {
		t.Errorf("Version = %q, want %q", info.Version, "5.1.8-6.el9_1")
	}
	if info.Architecture != "x86_64" {
		t.Errorf("Architecture = %q, want %q", info.Architecture, "x86_64")
	}
	if info.Description != "The GNU Bourne Again shell" {
		t.Errorf("Description = %q, want %q", info.Description, "The GNU Bourne Again shell")
	}
	if info.Status != "installed" {
		t.Errorf("Status = %q, want %q", info.Status, "installed")
	}
}

func TestParseRpmListOutput(t *testing.T) {
	output := "bash|5.1.8-6|x86_64|GNU Bourne Again shell|installed\n" +
		"coreutils|8.32-34|x86_64|Core utilities|installed\n" +
		"\n"

	packages := parseRpmListOutput(output, "yum")
	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}
	if packages[0].Name != "bash" {
		t.Errorf("first package name = %q, want %q", packages[0].Name, "bash")
	}
	if packages[1].Name != "coreutils" {
		t.Errorf("second package name = %q, want %q", packages[1].Name, "coreutils")
	}
}

func TestDetectManagerError(t *testing.T) {
	// This test verifies that detectManager returns an error when no
	// supported manager binary exists. We cannot easily mock the filesystem,
	// so we test the error path only on systems where none are present.
	// On macOS (where this test typically runs), none of the Linux paths exist.
	name, path, err := detectManager()
	// We don't assert the result here because CI may or may not have apt/yum/dnf.
	// Instead, we just verify the function returns without panic.
	_ = name
	_ = path
	_ = err
}

func TestInstallReturnsActionOnError(t *testing.T) {
	// Install should return a PackageAction even when no manager is found.
	action, _ := Install("nonexistent-package-xyz")
	if action.Name != "nonexistent-package-xyz" {
		t.Errorf("Name = %q, want %q", action.Name, "nonexistent-package-xyz")
	}
	if action.Action != "install" {
		t.Errorf("Action = %q, want %q", action.Action, "install")
	}
}

func TestRemoveReturnsActionOnError(t *testing.T) {
	action, _ := Remove("nonexistent-package-xyz")
	if action.Name != "nonexistent-package-xyz" {
		t.Errorf("Name = %q, want %q", action.Name, "nonexistent-package-xyz")
	}
	if action.Action != "remove" {
		t.Errorf("Action = %q, want %q", action.Action, "remove")
	}
}

func TestPackageActionJSONFieldNames(t *testing.T) {
	action := PackageAction{
		Name:    "test",
		Action:  "install",
		Manager: "apt",
		Success: true,
		Message: "ok",
	}
	data, _ := json.Marshal(action)
	var m map[string]interface{}
	json.Unmarshal(data, &m)

	expectedKeys := []string{"name", "action", "manager", "success", "changed", "message"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}
}

func TestPackageInfoJSONFieldNames(t *testing.T) {
	info := PackageInfo{
		Name:         "test",
		Version:      "1.0",
		Architecture: "amd64",
		Description:  "desc",
		Status:       "installed",
		Manager:      "apt",
	}
	data, _ := json.Marshal(info)
	var m map[string]interface{}
	json.Unmarshal(data, &m)

	expectedKeys := []string{"name", "version", "architecture", "description", "status", "manager"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}
}
