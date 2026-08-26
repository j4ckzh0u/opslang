package opsspec_test

import (
	"testing"

	"github.com/j4ckzh0u/opslang/internal/compiler"
	"github.com/j4ckzh0u/opslang/internal/interpreter"
	"github.com/j4ckzh0u/opslang/internal/opsspec"
	"github.com/j4ckzh0u/opslang/internal/runner"
)

// The three execution engines (interpreter, runner registry, AOT codegen)
// must expose exactly the canonical function set from opsspec. They used to
// drift silently: scripts that ran locally failed remotely with "unknown
// operation" and vice versa. This test pins the contract.
func TestEnginesExposeCanonicalFunctionSet(t *testing.T) {
	bridge := toSet(interpreter.SDKBuiltinNames())
	registry := toSet(runner.NewRegistry().ListOperations())
	codegen := toSet(compiler.SDKMappingNames())

	all := toSet(opsspec.Names(nil))
	controllerOnly := toSet(opsspec.Names(availPtr(opsspec.ControllerOnly)))

	for _, f := range opsspec.Funcs {
		// Interpreter (controller): everything.
		if !bridge[f.Name] {
			t.Errorf("interpreter bridge is missing %q", f.Name)
		}

		if f.Avail == opsspec.ControllerOnly {
			// Remote engines must NOT expose controller-only fan-out ops.
			if registry[f.Name] {
				t.Errorf("runner registry must not expose controller-only %q", f.Name)
			}
			continue
		}

		// Remote engines: everything else.
		if !registry[f.Name] {
			t.Errorf("runner registry is missing %q", f.Name)
		}
		if !codegen[f.Name] {
			t.Errorf("AOT codegen mapping is missing %q", f.Name)
		}
	}

	// No engine may expose a controller-only op remotely.
	for name := range controllerOnly {
		if registry[name] {
			t.Errorf("runner registry exposes controller-only %q", name)
		}
		if codegen[name] {
			t.Errorf("AOT codegen exposes controller-only %q", name)
		}
	}

	// Remote engines must not invent canonical-looking names that are not
	// in the spec (aliases and builtin VM ops are fine).
	builtins := toSet(opsspec.BuiltinOps)
	for name := range registry {
		if !all[name] && !isAlias(name) && !builtins[name] {
			t.Errorf("runner registry has non-canonical, non-alias op %q", name)
		}
	}
	for name := range codegen {
		if !all[name] && !isAlias(name) {
			t.Errorf("AOT codegen has non-canonical, non-alias mapping %q", name)
		}
	}
}

// Aliases must resolve to the canonical op in every engine that supports
// them, so existing instruction packages keep working.
func TestAliasesResolveToCanonical(t *testing.T) {
	r := runner.NewRegistry()
	for alias, canonical := range opsspec.Aliases {
		if _, ok := r.Get(alias); !ok {
			t.Errorf("runner registry does not resolve alias %q", alias)
		}
		if _, ok := r.Get(canonical); !ok {
			t.Errorf("runner registry is missing canonical %q behind alias", canonical)
		}
	}
}

// Argument-name signatures must agree between the spec, the generator and
// the registry helpers.
func TestArgNamesAgreeWithSpec(t *testing.T) {
	for _, f := range opsspec.Funcs {
		if f.Avail == opsspec.ControllerOnly {
			continue
		}
		names, ok := opsspec.ArgNames(f.Name)
		if !ok {
			t.Fatalf("ArgNames missing for %q", f.Name)
		}
		if len(names) != len(f.Args) {
			t.Errorf("%s: ArgNames() = %v, spec args = %v", f.Name, names, f.Args)
		}
		for i := range names {
			if names[i] != f.Args[i] {
				t.Errorf("%s: arg %d name mismatch: %q vs %q", f.Name, i, names[i], f.Args[i])
			}
		}
	}
}

func TestSpecTableIsUnique(t *testing.T) {
	seen := make(map[string]bool, len(opsspec.Funcs))
	for _, f := range opsspec.Funcs {
		if seen[f.Name] {
			t.Errorf("duplicate function %q in spec", f.Name)
		}
		seen[f.Name] = true
	}
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, i := range items {
		m[i] = true
	}
	return m
}

func isAlias(name string) bool {
	_, ok := opsspec.Aliases[name]
	return ok
}

func availPtr(a opsspec.Availability) *opsspec.Availability { return &a }
