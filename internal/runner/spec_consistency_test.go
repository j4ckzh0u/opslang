package runner

import (
	"testing"

	"github.com/j4ckzh0u/opslang/internal/opsspec"
)

// The runner is the remote half of every deploy. If the canonical opsspec
// table promises an op for all engines but the runner registry does not
// implement it, deploy mode fails on the host while `opsctl run` works
// locally — the worst kind of drift. This test keeps the registry glued to
// the spec table.
func TestRegistryCoversSpec(t *testing.T) {
	r := NewRegistry()
	for _, f := range opsspec.Funcs {
		if f.Avail == opsspec.ControllerOnly {
			continue
		}
		if !r.Has(f.Name) {
			t.Errorf("spec/runner mismatch: %s is in opsspec but not registered in the runner registry", f.Name)
		}
	}
}

// The ensure-family convergence ops are the Ansible-parity core; lock them
// against accidental unregistration and argument-name drift.
func TestEnsureFamilyRegistered(t *testing.T) {
	r := NewRegistry()
	for _, name := range []string{
		"pkg.ensure",
		"user.ensure", "user.absent",
		"group.ensure", "group.absent",
		"service.ensure", "service.ensure_enabled",
		"file.ensure",
	} {
		if !r.Has(name) {
			t.Errorf("runner registry must implement %s", name)
		}
	}

	wantArgs := map[string][]string{
		"pkg.ensure":             {"name"},
		"user.ensure":            {"username", "opts"},
		"user.absent":            {"username", "remove_home"},
		"group.ensure":           {"name", "opts"},
		"group.absent":           {"name"},
		"service.ensure":         {"name", "state"},
		"service.ensure_enabled": {"name", "enabled"},
		"file.ensure":            {"path", "state", "mode"},
	}
	for _, f := range opsspec.Funcs {
		if want, ok := wantArgs[f.Name]; ok {
			got := f.Args
			if len(got) != len(want) {
				t.Errorf("%s args = %v, want %v", f.Name, got, want)
				continue
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("%s args = %v, want %v", f.Name, got, want)
					break
				}
			}
		}
	}
}
