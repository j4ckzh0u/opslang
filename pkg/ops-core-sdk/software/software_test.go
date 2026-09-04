package software

import "testing"

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
