package runner

import (
	"fmt"
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/parser"
)

// ============================================================
// Helpers
// ============================================================

// mustParse parses source and returns the program, failing the test on error.
func mustParse(t *testing.T, source string) *ast.Program {
	t.Helper()
	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return prog
}

// findTask returns the first TaskStatement in the program, or fails the test.
func findTask(t *testing.T, prog *ast.Program) *ast.TaskStatement {
	t.Helper()
	for _, stmt := range prog.Statements {
		if task, ok := stmt.(*ast.TaskStatement); ok {
			return task
		}
	}
	t.Fatal("no task statement found in program")
	return nil
}

// generate parses source, finds its task, and generates instructions.
// The generator defaults to read_only, matching undeclared scripts.
func generate(t *testing.T, source string) *InstructionPackage {
	t.Helper()
	return generateWithPrivilege(t, source, "")
}

// generateAsAdmin is generate with admin privilege, for tests whose task
// bodies legitimately perform mutating operations.
func generateAsAdmin(t *testing.T, source string) *InstructionPackage {
	t.Helper()
	return generateWithPrivilege(t, source, ast.PrivilegeAdmin)
}

// generateWithPrivilege parses source, finds its task, and generates
// instructions under the given privilege level.
func generateWithPrivilege(t *testing.T, source string, priv ast.PrivilegeLevel) *InstructionPackage {
	t.Helper()
	task := findTask(t, mustParse(t, source))
	gen := &InstructionGenerator{Privilege: priv}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	return pkg
}

// generateErr parses source and expects Generate to fail with a message
// containing want.
func generateErr(t *testing.T, source, want string) {
	t.Helper()
	task := findTask(t, mustParse(t, source))
	gen := &InstructionGenerator{}
	_, err := gen.Generate(task, false)
	if err == nil {
		t.Fatalf("expected Generate error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected Generate error containing %q, got: %v", want, err)
	}
}

// ============================================================
// Generate tests
// ============================================================

func TestGenerate_SimpleTaskWithCPUAndReport(t *testing.T) {
	pkg := generate(t, `task "check" on "host1" {
		let cpu = sys.cpu.usage()
		report { cpu: cpu }
	}`)

	if pkg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", pkg.Version)
	}
	if pkg.TaskID == "" {
		t.Error("expected non-empty task ID")
	}
	if pkg.DryRun {
		t.Error("expected dry_run to be false")
	}
	if len(pkg.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(pkg.Instructions))
	}

	inst0 := pkg.Instructions[0]
	if inst0.Op != "sys.cpu.usage" {
		t.Errorf("expected op sys.cpu.usage, got %s", inst0.Op)
	}
	if inst0.Assign != "cpu" {
		t.Errorf("expected assign cpu, got %s", inst0.Assign)
	}

	inst1 := pkg.Instructions[1]
	if inst1.Op != "report" {
		t.Errorf("expected op report, got %s", inst1.Op)
	}
	// Variable references must be explicit "$name" markers.
	if cpuRef, ok := inst1.Args["cpu"]; !ok || cpuRef != "$cpu" {
		t.Errorf("expected report arg cpu=$cpu, got %v", inst1.Args)
	}
}

func TestGenerate_LetAssignments(t *testing.T) {
	pkg := generate(t, `task "test" on "host1" {
		let x = 42
		let name = "hello"
		let flag = true
	}`)

	if len(pkg.Instructions) != 3 {
		t.Fatalf("expected 3 instructions, got %d", len(pkg.Instructions))
	}

	inst0 := pkg.Instructions[0]
	if inst0.Op != "set" {
		t.Errorf("inst0: expected op set, got %s", inst0.Op)
	}
	if inst0.Assign != "x" {
		t.Errorf("inst0: expected assign x, got %s", inst0.Assign)
	}
	if val, ok := inst0.Args["value"]; !ok || val != int64(42) {
		t.Errorf("inst0: expected value=42, got %v (%T)", val, val)
	}

	inst1 := pkg.Instructions[1]
	if inst1.Assign != "name" || inst1.Args["value"] != "hello" {
		t.Errorf("inst1: got %+v", inst1)
	}

	inst2 := pkg.Instructions[2]
	if inst2.Assign != "flag" || inst2.Args["value"] != true {
		t.Errorf("inst2: got %+v", inst2)
	}
}

func TestGenerate_FileOperations(t *testing.T) {
	pkg := generate(t, `task "files" on "host1" {
		let content = file.read("/etc/hosts")
		file.exists("/tmp/test")
	}`)

	if len(pkg.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(pkg.Instructions))
	}

	inst0 := pkg.Instructions[0]
	if inst0.Op != "file.read" || inst0.Assign != "content" {
		t.Errorf("inst0: %+v", inst0)
	}
	if path, ok := inst0.Args["path"]; !ok || path != "/etc/hosts" {
		t.Errorf("expected path=/etc/hosts, got %v", path)
	}

	inst1 := pkg.Instructions[1]
	if inst1.Op != "file.exists" || inst1.Assign != "" {
		t.Errorf("inst1: %+v", inst1)
	}
	if path, ok := inst1.Args["path"]; !ok || path != "/tmp/test" {
		t.Errorf("expected path=/tmp/test, got %v", path)
	}
}

func TestGenerate_ArgumentNamesAreCanonical(t *testing.T) {
	// The generated arg keys must match the op signatures the registry
	// validates — "arg0"-style keys silently produced empty paths before.
	pkg := generateAsAdmin(t, `task "t" on "h" {
		file.copy("/src", "/dst")
		net.tcp_check("localhost", 22)
		file.checksum("/etc/hosts", "md5")
	}`)

	if pkg.Instructions[0].Args["src"] != "/src" || pkg.Instructions[0].Args["dst"] != "/dst" {
		t.Errorf("file.copy args: %+v", pkg.Instructions[0].Args)
	}
	tcp := pkg.Instructions[1].Args
	if tcp["host"] != "localhost" || tcp["port"] != int64(22) {
		t.Errorf("net.tcp_check args: %+v", tcp)
	}
	sum := pkg.Instructions[2].Args
	if sum["path"] != "/etc/hosts" || sum["algo"] != "md5" {
		t.Errorf("file.checksum args: %+v", sum)
	}
}

func TestGenerate_UnknownFunctionFails(t *testing.T) {
	generateErr(t, `task "t" on "h" {
		sys.nonexistent_function()
	}`, "unknown function")
}

func TestGenerate_TooManyArgumentsFails(t *testing.T) {
	generateErr(t, `task "t" on "h" {
		file.read("/a", "/b", "/c")
	}`, "at most")
}

func TestGenerate_DryRunFlag(t *testing.T) {
	task := findTask(t, mustParse(t, `task "test" on "host1" {
		sys.cpu.usage()
	}`))
	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, true)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if !pkg.DryRun {
		t.Error("expected dry_run to be true")
	}
}

func TestGenerate_ReportWithSubCalls(t *testing.T) {
	pkg := generate(t, `task "test" on "host1" {
		report { host: sys.hostname(), cpu: sys.cpu.usage() }
	}`)

	if len(pkg.Instructions) != 3 {
		t.Fatalf("expected 3 instructions, got %d", len(pkg.Instructions))
	}

	if pkg.Instructions[0].Op != "sys.hostname" || pkg.Instructions[0].Assign != "__tmp_0" {
		t.Errorf("inst0: %+v", pkg.Instructions[0])
	}
	if pkg.Instructions[1].Op != "sys.cpu.usage" || pkg.Instructions[1].Assign != "__tmp_1" {
		t.Errorf("inst1: %+v", pkg.Instructions[1])
	}

	report := pkg.Instructions[2]
	if report.Op != "report" {
		t.Fatalf("inst2: expected report, got %s", report.Op)
	}
	if report.Args["host"] != "$__tmp_0" {
		t.Errorf("expected host=$__tmp_0, got %v", report.Args["host"])
	}
	if report.Args["cpu"] != "$__tmp_1" {
		t.Errorf("expected cpu=$__tmp_1, got %v", report.Args["cpu"])
	}
}

func TestGenerate_AlertStatement(t *testing.T) {
	pkg := generate(t, `task "test" on "host1" {
		alert("something went wrong")
	}`)

	if len(pkg.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(pkg.Instructions))
	}
	inst := pkg.Instructions[0]
	if inst.Op != "alert" || inst.Args["message"] != "something went wrong" {
		t.Errorf("got %+v", inst)
	}
}

func TestGenerate_PrintBuiltin(t *testing.T) {
	pkg := generate(t, `task "test" on "host1" {
		print("hello world")
	}`)

	if len(pkg.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(pkg.Instructions))
	}
	inst := pkg.Instructions[0]
	if inst.Op != "log" || inst.Args["message"] != "hello world" {
		t.Errorf("got %+v", inst)
	}
}

// Control flow must HARD FAIL, not silently degrade. The old generator
// executed if-bodies unconditionally and logged ensure as "check done".
func TestGenerate_ControlFlowFails(t *testing.T) {
	generateErr(t, `task "t" on "h" {
		if true { print("x") }
	}`, "if statement")

	generateErr(t, `task "t" on "h" {
		for let i = 0; i < 10; i = i + 1 { print("loop") }
	}`, "for loop")

	generateErr(t, `task "t" on "h" {
		while false { print("loop") }
	}`, "while loop")

	generateErr(t, `task "t" on "h" {
		fn helper() { print("hi") }
	}`, "function definition")

	generateErr(t, `task "t" on "h" {
		ensure file.exists("/tmp/x").exists { file.mkdir("/tmp/x") }
	}`, "ensure statement")

	generateErr(t, `task "t" on "h" {
		parallel { sys.cpu.usage() }
	}`, "parallel block")
}

// Binary/unary/member expressions cannot be evaluated by the linear runner.
func TestGenerate_ComputedExpressionFails(t *testing.T) {
	generateErr(t, `task "t" on "h" {
		let x = 1 + 2
	}`, "cannot evaluate")

	generateErr(t, `task "t" on "h" {
		let name = sys.hostname().hostname
	}`, "cannot dereference")
}

func TestGenerateFromStatements(t *testing.T) {
	prog := mustParse(t, `let x = sys.cpu.usage()
report { x: x }`)

	gen := &InstructionGenerator{}
	pkg, err := gen.GenerateFromStatements(prog.Statements, false)
	if err != nil {
		t.Fatalf("GenerateFromStatements failed: %v", err)
	}
	if len(pkg.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(pkg.Instructions))
	}
}

func TestGenerate_EmptyTask(t *testing.T) {
	pkg := generate(t, `task "test" on "host1" {}`)
	if len(pkg.Instructions) != 0 {
		t.Errorf("expected 0 instructions, got %d", len(pkg.Instructions))
	}
}

func TestGenerate_TaskIDUniqueness(t *testing.T) {
	pkg1 := generate(t, `task "test" on "host1" { sys.cpu.usage() }`)
	pkg2 := generate(t, `task "test" on "host1" { sys.cpu.usage() }`)
	if pkg1.TaskID == pkg2.TaskID {
		t.Error("expected unique task IDs across generations")
	}
}

func TestGenerate_CopyFileOperation(t *testing.T) {
	pkg := generateAsAdmin(t, `task "test" on "host1" {
		file.copy("/src/file", "/dst/file")
	}`)

	if len(pkg.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(pkg.Instructions))
	}
	inst := pkg.Instructions[0]
	if inst.Op != "file.copy" {
		t.Errorf("expected op file.copy, got %s", inst.Op)
	}
	if inst.Args["src"] != "/src/file" || inst.Args["dst"] != "/dst/file" {
		t.Errorf("args: %+v", inst.Args)
	}
}

// The generated package must pass registry validation end to end: every
// emitted op must exist and every instruction must be executable.
func TestGenerate_PackageValidatesAgainstRegistry(t *testing.T) {
	pkg := generate(t, `task "t" on "h" {
		let info = sys.os()
		let hosts = file.read("/etc/hosts")
		print("done")
		alert("check")
		report { os: info, hosts: hosts }
	}`)
	if err := ValidatePackage(pkg); err != nil {
		t.Fatalf("generated package failed validation: %v", err)
	}
}

// A script with a dry-run-safe pipeline executed through Run() must
// actually resolve variable references between instructions.
func TestGenerateAndRun_VariableFlow(t *testing.T) {
	pkg := generate(t, `task "t" on "h" {
		let data = json.encode([1, 2, 3])
		report { encoded: data }
	}`)

	out := Run(pkg, NewRegistry())
	if out.Status != "ok" {
		t.Fatalf("status = %q, errors = %v", out.Status, out.Errors)
	}
	encoded, ok := out.Data["encoded"]
	if !ok {
		t.Fatalf("report data missing 'encoded': %+v", out.Data)
	}
	// json.encode returns a map with the encoded string under a field.
	if encoded == nil {
		t.Fatal("encoded is nil: variable reference not resolved")
	}
}

// TestGenerate_DataBuiltinsRefused pins the honest-refusal contract: the
// linear runner VM has no expression evaluator, so language-level data
// builtins (split/join/sort/...) must fail generation with a clear error,
// never a silent mistranslation. Scripts using them deploy in AOT mode.
func TestGenerate_DataBuiltinsRefused(t *testing.T) {
	for _, fn := range []string{"split", "join", "replace", "upper", "lower",
		"trim", "contains", "index_of", "sort", "reverse", "keys", "values"} {
		src := fmt.Sprintf(`task "check" on "host1" {
			let parts = %s("a,b", ",")
			report { p: parts }
		}`, fn)
		task := findTask(t, mustParse(t, src))
		gen := &InstructionGenerator{}
		if _, err := gen.Generate(task, false); err == nil {
			t.Errorf("%s(): expected generation to refuse the builtin, got success", fn)
		}
	}
}

// TestGenerate_MetaBuiltinsRefused pins the same honest-refusal contract
// for introspection builtins: doc/ops read the controller-side opsspec
// table, which a remote linear runner VM does not carry.
func TestGenerate_MetaBuiltinsRefused(t *testing.T) {
	for _, fn := range []string{"doc", "ops"} {
		src := fmt.Sprintf(`task "check" on "host1" {
			let names = %s("sys.")
			report { n: names }
		}`, fn)
		task := findTask(t, mustParse(t, src))
		gen := &InstructionGenerator{}
		if _, err := gen.Generate(task, false); err == nil {
			t.Errorf("%s(): expected generation to refuse the meta builtin, got success", fn)
		}
	}
}

// TestGenerate_ParallelForRefused pins the honest-refusal contract for
// the fan-out loop in runner mode.
func TestGenerate_ParallelForRefused(t *testing.T) {
	src := `task "check" on "host1" {
	parallel for h in ["a", "b"] {
		let r = ping(h)
		report { r: r }
	}
}`
	task := findTask(t, mustParse(t, src))
	gen := &InstructionGenerator{}
	if _, err := gen.Generate(task, false); err == nil {
		t.Error("parallel for should be refused in runner mode, got success")
	}
}
