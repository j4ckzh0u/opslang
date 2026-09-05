package software

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExecutableDirectory(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "unix", path: "/usr/bin/nginx", want: "/usr/bin"},
		{name: "windows", path: `C:\Program Files\App\app.exe`, want: `C:\Program Files\App`},
		{name: "bare", path: "app", want: ""},
		{name: "empty", path: "", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := executableDirectory(test.path); got != test.want {
				t.Fatalf("executableDirectory(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

func TestInventoryInitializesSlices(t *testing.T) {
	result := InventoryResult{Packages: make([]Package, 0), RunningPrograms: make([]RunningProgram, 0), Errors: make([]Error, 0)}
	if result.Packages == nil || result.RunningPrograms == nil || result.Errors == nil {
		t.Fatal("inventory result slices must be initialized")
	}
}

func TestEnrichProgramPackages(t *testing.T) {
	programs := []RunningProgram{{Executable: "/usr/bin/example"}, {Executable: "/opt/example/bin/app"}, {Executable: ""}}
	packages := []Package{
		{Name: "example", Version: "1.2.3", InstalledFiles: []string{"/usr/bin/example"}},
		{Name: "example-app", Version: "4.5.6", InstallLocation: "/opt/example"},
	}
	enrichProgramPackages(programs, packages)
	if programs[0].PackageName != "example" || programs[0].Version != "1.2.3" {
		t.Fatalf("exact file match was not applied: %+v", programs[0])
	}
	if programs[1].PackageName != "example-app" || programs[1].PackageVersion != "4.5.6" {
		t.Fatalf("install location match was not applied: %+v", programs[1])
	}
	if programs[2].PackageName != "" || programs[2].Version != "" {
		t.Fatalf("unmatched process was incorrectly enriched: %+v", programs[2])
	}
}

func TestDpkgPackageFilesUsesArchitectureManifest(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "example:amd64.list")
	if err := os.WriteFile(manifest, []byte("/usr/bin/example\n\n/usr/share/doc/example\n"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	got, err := dpkgPackageFiles(dir, "example", "amd64")
	if err != nil {
		t.Fatalf("dpkgPackageFiles: %v", err)
	}
	want := []string{"/usr/bin/example", "/usr/share/doc/example"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dpkgPackageFiles() = %#v, want %#v", got, want)
	}
}

func TestPackageFilesRejectsEmptyName(t *testing.T) {
	if _, err := packageFiles("apt", "", "amd64"); err == nil {
		t.Fatal("packageFiles must reject an empty package name")
	}
}

func TestParseRPMFileListOutputGroupsPackages(t *testing.T) {
	output := "bash|x86_64|/usr/bin/bash\n" +
		"bash|x86_64|/usr/share/doc/bash\n" +
		"curl|x86_64|/usr/bin/curl\n" +
		"invalid line\n"

	got := parseRPMFileListOutput(output)
	if want := []string{"/usr/bin/bash", "/usr/share/doc/bash"}; !reflect.DeepEqual(got[rpmPackageKey("bash", "x86_64")], want) {
		t.Fatalf("bash files = %#v, want %#v", got[rpmPackageKey("bash", "x86_64")], want)
	}
	if want := []string{"/usr/bin/curl"}; !reflect.DeepEqual(got[rpmPackageKey("curl", "x86_64")], want) {
		t.Fatalf("curl files = %#v, want %#v", got[rpmPackageKey("curl", "x86_64")], want)
	}
}

func TestParseRPMFileListOutputHandlesEmptyInput(t *testing.T) {
	got := parseRPMFileListOutput("")
	if got == nil {
		t.Fatal("empty RPM output must return an initialized map")
	}
	if len(got) != 0 {
		t.Fatalf("empty RPM output produced %d packages", len(got))
	}
}
