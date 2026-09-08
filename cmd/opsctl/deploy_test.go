package main

import (
	"strings"
	"testing"
	"time"

	"github.com/j4ckzh0u/opslang/internal/ast"
	opsexec "github.com/j4ckzh0u/opslang/internal/exec"
	"github.com/j4ckzh0u/opslang/internal/parser"
	"github.com/j4ckzh0u/opslang/internal/runner"
	"github.com/j4ckzh0u/opslang/internal/security"
)

func TestAOTInstructionPackageUsesTransportPrivilege(t *testing.T) {
	pkg := newAOTInstructionPackage("task-1", false, ast.PrivilegeReadOnly)
	if pkg.Privilege != string(ast.PrivilegeAdmin) {
		t.Fatalf("AOT transport package privilege = %q, want admin", pkg.Privilege)
	}
	if len(pkg.Instructions) != 1 || pkg.Instructions[0].Op != "binary.exec" {
		t.Fatalf("unexpected AOT package: %+v", pkg.Instructions)
	}
}

func TestDefaultBuildApprovalGateBlocksUnapprovedProductionAdmin(t *testing.T) {
	prog := parseProgram(t, `privilege: admin
file.write("/tmp/approval-check", "data")`)
	targets := []opsexec.Target{{Name: "prod-01", Host: "10.0.0.1", Tags: map[string]string{"env": "prod"}}}
	oldInteractive, oldConfirm := deployStdinIsInteractive, deployConfirmFn
	deployStdinIsInteractive = func() bool { return false }
	deployConfirmFn = func(string) bool { return false }
	t.Cleanup(func() {
		deployStdinIsInteractive, deployConfirmFn = oldInteractive, oldConfirm
	})
	if _, err := enforceDeployApprovalGate("test.ops", prog, targets, false, approveNonInteractive); err == nil {
		t.Fatal("default build must block unapproved privileged production deploy")
	}
}

func parseProgram(t *testing.T, source string) *ast.Program {
	t.Helper()
	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return prog
}

func testTargets() []opsexec.Target {
	return []opsexec.Target{
		{Name: "web-01", Host: "10.0.0.1", Port: 22, User: "root"},
		{Name: "web-02", Host: "10.0.0.2", Port: 22, User: "root"},
		{Name: "db-01", Host: "10.0.0.10", Port: 22, User: "root"},
	}
}

func TestResolveDeployMode(t *testing.T) {
	linear := parseProgram(t, `let cpu = sys.cpu.usage()
report { cpu: cpu }`)
	if got := resolveDeployMode("auto", linear); got != "runner" {
		t.Errorf("linear script: mode = %q, want runner", got)
	}

	controlFlow := parseProgram(t, `if true { print("x") }`)
	if got := resolveDeployMode("auto", controlFlow); got != "aot" {
		t.Errorf("control flow script: mode = %q, want aot", got)
	}

	if got := resolveDeployMode("runner", controlFlow); got != "runner" {
		t.Errorf("explicit runner override failed: %q", got)
	}
}

func TestBuildDeploySteps_PreludeAndTaskRouting(t *testing.T) {
	prog := parseProgram(t, `
print("global setup")
task "web" on "web-*" {
	sys.cpu.usage()
}
task "db" on "db-01" {
	sys.memory.info()
}
`)

	steps, err := buildDeploySteps(prog, testTargets(), "task-1", security.GetScriptPrivilege(prog))
	if err != nil {
		t.Fatalf("buildDeploySteps failed: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps (prelude + 2 tasks), got %d", len(steps))
	}
	if steps[0].name != "main" || len(steps[0].targets) != 3 {
		t.Errorf("prelude step: name=%q targets=%d", steps[0].name, len(steps[0].targets))
	}
	if steps[1].name != "web" || len(steps[1].targets) != 2 {
		t.Errorf("web step: name=%q targets=%d, want 2", steps[1].name, len(steps[1].targets))
	}
	for _, tgt := range steps[1].targets {
		if !strings.HasPrefix(tgt.Name, "web-") {
			t.Errorf("web step wrongly routed to %q", tgt.Name)
		}
	}
	if steps[2].name != "db" || len(steps[2].targets) != 1 || steps[2].targets[0].Name != "db-01" {
		t.Errorf("db step routing wrong: %+v", steps[2].targets)
	}
	for _, step := range steps {
		if step.main.TaskID == "" {
			t.Errorf("step %q has empty TaskID", step.name)
		}
	}
}

func TestBuildDeploySteps_GeneratesIndependentRecoveryPackages(t *testing.T) {
	prog := parseProgram(t, `privilege: admin
task "deploy" on "web-01" {
	file.write("/tmp/app", "new")
} rescue {
	file.write("/tmp/app", "old")
} always {
	file.exists("/tmp/app")
}`)

	steps, err := buildDeploySteps(prog, testTargets(), "task-1", security.GetScriptPrivilege(prog))
	if err != nil {
		t.Fatalf("buildDeploySteps failed: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("steps = %d, want 1", len(steps))
	}
	step := steps[0]
	if step.main == nil || step.rescue == nil || step.always == nil {
		t.Fatalf("missing phase package: %+v", step)
	}
	if step.main.TaskID != "task-1-deploy" || step.rescue.TaskID != "task-1-deploy-rescue" || step.always.TaskID != "task-1-deploy-always" {
		t.Fatalf("unexpected phase task IDs: %q %q %q", step.main.TaskID, step.rescue.TaskID, step.always.TaskID)
	}
	for _, pkg := range []*runner.InstructionPackage{step.main, step.rescue, step.always} {
		if pkg.Privilege != "admin" {
			t.Fatalf("phase privilege = %q, want admin", pkg.Privilege)
		}
	}
}

func TestPhaseStatus(t *testing.T) {
	success := &opsexec.HostResult{Status: "success"}
	failed := &opsexec.HostResult{Status: "failed"}
	tests := []struct {
		name   string
		phases *deployHostPhases
		want   string
	}{
		{name: "missing phases", phases: nil, want: "failed"},
		{name: "success", phases: &deployHostPhases{Main: success}, want: "success"},
		{name: "failed without rescue", phases: &deployHostPhases{Main: failed}, want: "failed"},
		{name: "rolled back", phases: &deployHostPhases{Main: failed, Rescue: success}, want: "rolled_back"},
		{name: "rollback failed", phases: &deployHostPhases{Main: failed, Rescue: failed}, want: "rollback_failed"},
		{name: "cleanup has priority", phases: &deployHostPhases{Main: failed, Rescue: failed, Always: failed}, want: "cleanup_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phaseStatus(tt.phases); got != tt.want {
				t.Fatalf("phaseStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTargetsWithFailedResultsIncludesMissingAndFailed(t *testing.T) {
	targets := testTargets()
	summary := &opsexec.Summary{Results: map[string]*opsexec.HostResult{
		"web-01": {Status: "success"},
		"web-02": {Status: "failed"},
	}}
	failed := targetsWithFailedResults(targets, summary)
	if len(failed) != 2 || failed[0].Name != "web-02" || failed[1].Name != "db-01" {
		t.Fatalf("failed targets = %+v, want web-02 and missing db-01", failed)
	}
}

func TestRunDeployStep_RollsBackOnlyFailedHostsAndAlwaysRuns(t *testing.T) {
	targets := testTargets()[:2]
	step := deployStep{
		name: "deploy", targets: targets,
		main: &runner.InstructionPackage{}, rescue: &runner.InstructionPackage{}, always: &runner.InstructionPackage{},
	}
	calls := make([]string, 0, 3)
	execute := func(phase string, _ *runner.InstructionPackage, phaseTargets []opsexec.Target) *opsexec.Summary {
		calls = append(calls, phase)
		results := make(map[string]*opsexec.HostResult)
		for _, target := range phaseTargets {
			status := "success"
			if phase == "main" && target.Name == "web-02" {
				status = "failed"
			}
			results[target.Name] = &opsexec.HostResult{Status: status}
		}
		if phase == "rescue" && (len(phaseTargets) != 1 || phaseTargets[0].Name != "web-02") {
			t.Fatalf("rescue targets = %+v, want only web-02", phaseTargets)
		}
		if phase == "always" && len(phaseTargets) != 2 {
			t.Fatalf("always targets = %+v, want both started hosts", phaseTargets)
		}
		return &opsexec.Summary{Status: "success", Results: results}
	}
	agg := &deployAggregate{}
	if stopped := runDeployStep(step, false, execute, agg); !stopped {
		t.Fatal("main failure must stop later tasks")
	}
	if strings.Join(calls, ",") != "main,rescue,always" {
		t.Fatalf("phase order = %v", calls)
	}
	if got := agg.TaskResults["deploy"]["web-01"].Status; got != "success" {
		t.Fatalf("web-01 status = %q, want success", got)
	}
	if got := agg.TaskResults["deploy"]["web-02"].Status; got != "rolled_back" {
		t.Fatalf("web-02 status = %q, want rolled_back", got)
	}
}

func TestRunDeployStep_DryRunPreviewsRescueOnAllHosts(t *testing.T) {
	targets := testTargets()[:2]
	step := deployStep{name: "preview", targets: targets, main: &runner.InstructionPackage{}, rescue: &runner.InstructionPackage{}}
	execute := func(phase string, _ *runner.InstructionPackage, phaseTargets []opsexec.Target) *opsexec.Summary {
		if phase == "rescue" && len(phaseTargets) != len(targets) {
			t.Fatalf("dry-run rescue targets = %d, want %d", len(phaseTargets), len(targets))
		}
		results := make(map[string]*opsexec.HostResult)
		for _, target := range phaseTargets {
			results[target.Name] = &opsexec.HostResult{Status: "success"}
		}
		return &opsexec.Summary{Status: "success", Results: results}
	}
	agg := &deployAggregate{}
	if stopped := runDeployStep(step, true, execute, agg); stopped {
		t.Fatal("successful dry-run must continue")
	}
	for _, target := range targets {
		if agg.TaskResults["preview"][target.Name].Rescue == nil {
			t.Fatalf("host %q has no rescue preview", target.Name)
		}
	}
}

func TestBuildDeploySteps_TaskMatchingNothingFails(t *testing.T) {
	prog := parseProgram(t, `task "x" on "cache-*" { sys.cpu.usage() }`)
	_, err := buildDeploySteps(prog, testTargets(), "task-1", security.GetScriptPrivilege(prog))
	if err == nil {
		t.Fatal("task selecting no targets must fail the deploy, not silently skip")
	}
	if !strings.Contains(err.Error(), "selects none") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildDeploySteps_ControlFlowInTaskFails(t *testing.T) {
	prog := parseProgram(t, `task "x" on "web-01" {
	if true { sys.cpu.usage() }
}`)
	_, err := buildDeploySteps(prog, testTargets(), "task-1", security.GetScriptPrivilege(prog))
	if err == nil {
		t.Fatal("control flow in runner mode must fail generation")
	}
	if !strings.Contains(err.Error(), "aot") {
		t.Errorf("error should point at AOT mode: %v", err)
	}
}

func TestBuildDeploySteps_VariableTargetFails(t *testing.T) {
	prog := parseProgram(t, `let hosts = ["a"]
task "x" on hosts { sys.cpu.usage() }`)
	_, err := buildDeploySteps(prog, testTargets(), "task-1", security.GetScriptPrivilege(prog))
	if err == nil {
		t.Fatal("variable on-clause must fail at deploy time")
	}
}

func TestSelectTaskTargets_MatchForms(t *testing.T) {
	prog := parseProgram(t, `task "t" on "10.0.0.1" {}`)
	task := prog.Statements[0].(*ast.TaskStatement)
	selected, err := selectTaskTargets(task, testTargets())
	if err != nil {
		t.Fatalf("selectTaskTargets: %v", err)
	}
	if len(selected) != 1 || selected[0].Name != "web-01" {
		t.Errorf("host-address matching failed: %+v", selected)
	}

	prog = parseProgram(t, `task "t" on "root@10.0.0.2" {}`)
	task = prog.Statements[0].(*ast.TaskStatement)
	selected, err = selectTaskTargets(task, testTargets())
	if err != nil {
		t.Fatalf("selectTaskTargets: %v", err)
	}
	if len(selected) != 1 || selected[0].Name != "web-02" {
		t.Errorf("user@host matching failed: %+v", selected)
	}
}

func TestGenerateTaskID(t *testing.T) {
	id1 := generateTaskID("scripts/check.ops")
	id2 := generateTaskID("scripts/check.ops")
	if id1 == id2 {
		t.Error("expected unique task IDs")
	}
	if !strings.Contains(id1, "scripts_check") {
		t.Errorf("unexpected ID format: %q", id1)
	}
}

func TestOutputDeployResultPartialIsError(t *testing.T) {
	saved := deployOutput
	deployOutput = t.TempDir() + "/result.json"
	defer func() { deployOutput = saved }()

	agg := &deployAggregate{
		TaskID:  "t1",
		Status:  "partial",
		Targets: []string{"h1"},
		Results: map[string]*opsexec.HostResult{},
	}
	if err := outputDeployResult(agg, nowUTC(), "x.ops"); err == nil {
		t.Error("partial deploy must return an error (some hosts failed)")
	}

	agg.Status = "failed"
	if err := outputDeployResult(agg, nowUTC(), "x.ops"); err == nil {
		t.Error("failed deploy must return an error")
	}

	agg.Status = "success"
	if err := outputDeployResult(agg, nowUTC(), "x.ops"); err != nil {
		t.Errorf("successful deploy returned error: %v", err)
	}
}

func nowUTC() (t time.Time) { return time.Now().UTC() }

func TestBuildDeploySteps_ReadOnlyTaskWithMutatingCallFails(t *testing.T) {
	// Generation-time privilege enforcement on the runner path: the task
	// body's mutating call contradicts the script's read_only declaration,
	// so the deploy must fail on the controller, not on the host.
	prog := parseProgram(t, `privilege: read_only
task "x" on "web-01" {
	file.write("/tmp/never.txt", "nope")
}`)
	_, err := buildDeploySteps(prog, testTargets(), "task-1", security.GetScriptPrivilege(prog))
	if err == nil {
		t.Fatal("read_only task with file.write must fail generation")
	}
	if !strings.Contains(err.Error(), "privilege denied") || !strings.Contains(err.Error(), "file.write") {
		t.Errorf("error must name the denied function: %v", err)
	}
}

func TestBuildDeploySteps_AdminTaskWithMutatingCallSucceeds(t *testing.T) {
	prog := parseProgram(t, `privilege: admin
task "x" on "web-01" {
	file.write("/tmp/ok.txt", "yes")
}`)
	steps, err := buildDeploySteps(prog, testTargets(), "task-1", security.GetScriptPrivilege(prog))
	if err != nil {
		t.Fatalf("admin task with file.write must generate: %v", err)
	}
	for _, step := range steps {
		if step.main.Privilege != "admin" {
			t.Errorf("step %q package privilege = %q, want admin (runner second check relies on it)", step.name, step.main.Privilege)
		}
	}
}
