package main

// Every tracked example script must actually run with `opsctl run` and
// exit 0. Examples are the project's documentation-in-code; a broken
// example is a documented feature that does not work. This test walks
// examples/*.ops through the real interpreter so they can never silently
// rot again.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/opslang/opslang/internal/interpreter"
	"github.com/opslang/opslang/internal/parser"
	"github.com/opslang/opslang/internal/runner"
)

func TestExamplesAllRun(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	examplesDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "examples")

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("read examples dir: %v", err)
	}

	ran := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ops") {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(examplesDir, name)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			p := parser.New(string(source), path)
			prog, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			interp := interpreter.New(nil)
			interpreter.RegisterSDKBuiltins(interp)
			if _, err := interp.Execute(prog); err != nil {
				t.Fatalf("execution error: %v", err)
			}
		})
		ran++
	}

	if ran < 20 {
		t.Errorf("expected at least 20 example scripts, ran %d - did the examples dir move?", ran)
	}
}

// The deploy-mode instruction generator must accept every example that the
// auto mode would route to runner mode (linear scripts), and honestly
// reject the control-flow ones.
func TestExamplesRunnerModeRoutingIsConsistent(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	examplesDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "examples")

	entries, _ := os.ReadDir(examplesDir)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ops") {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(examplesDir, name)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			p := parser.New(string(source), path)
			prog, err := p.Parse()
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			mode := resolveDeployMode("auto", prog)
			if mode != "runner" {
				return // control flow -> aot, generator rejection not required
			}

			gen := &runner.InstructionGenerator{}
			if _, err := gen.GenerateFromStatements(prog.Statements, false); err != nil {
				t.Fatalf("auto mode picked runner for %q but generation failed: %v (mode selection and generator disagree)", name, err)
			}
		})
	}
}
