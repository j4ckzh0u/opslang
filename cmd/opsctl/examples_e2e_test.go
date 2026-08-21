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
	"time"

	"github.com/opslang/opslang/internal/ast"
	opsexec "github.com/opslang/opslang/internal/exec"
	"github.com/opslang/opslang/internal/interpreter"
	"github.com/opslang/opslang/internal/parser"
	"github.com/opslang/opslang/internal/runner"
	"github.com/opslang/opslang/internal/security"
)

// hasRoutedTask reports whether the program contains a task with an
// on-clause (deploy-only example).
func hasRoutedTask(prog *ast.Program) bool {
	for _, stmt := range prog.Statements {
		if task, ok := stmt.(*ast.TaskStatement); ok && task.Targets != nil {
			return true
		}
	}
	return false
}

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
			if name == "archive_and_download.ops" && os.Getenv("OPSLANG_RUN_NETWORK_EXAMPLES") != "1" {
				t.Skip("requires external network; set OPSLANG_RUN_NETWORK_EXAMPLES=1 to enable")
			}
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

			// Deploy examples (task with on-clauses) must route, not run
			// locally; everything else runs end to end.
			if hasRoutedTask(prog) {
				targets := []opsexec.Target{
					{Name: "web-01", Host: "10.0.0.1", User: "root"},
					{Name: "db-01", Host: "10.0.0.2", User: "root"},
				}
				steps, err := buildDeploySteps(prog, targets, "test", security.GetScriptPrivilege(prog))
				if err != nil {
					t.Fatalf("deploy example must build valid steps: %v", err)
				}
				if len(steps) == 0 {
					t.Fatal("deploy example produced no steps")
				}
				return
			}

			interp := interpreter.New(nil)
			interpreter.RegisterSDKBuiltins(interp)

			// Per-example timeout prevents a single hanging example
			// (network call, port wait, blocked stdin) from killing
			// the entire suite.
			done := make(chan error, 1)
			go func() {
				_, err := interp.Execute(prog)
				done <- err
			}()
			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("execution error: %v", err)
				}
			case <-time.After(10 * time.Second):
				t.Fatalf("example timed out after 10s (likely hanging on network/port wait)")
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
