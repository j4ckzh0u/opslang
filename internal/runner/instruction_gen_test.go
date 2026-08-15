package runner

import (
	"testing"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/parser"
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

// ============================================================
// Generate tests
// ============================================================

func TestGenerate_SimpleTaskWithCPUAndReport(t *testing.T) {
	source := `task "check" on "host1" {
		let cpu = sys.cpu.usage()
		report { cpu: cpu }
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

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

	// First instruction: sys.cpu.usage with assign "cpu"
	inst0 := pkg.Instructions[0]
	if inst0.Op != "sys.cpu.usage" {
		t.Errorf("expected op sys.cpu.usage, got %s", inst0.Op)
	}
	if inst0.Assign != "cpu" {
		t.Errorf("expected assign cpu, got %s", inst0.Assign)
	}

	// Second instruction: report
	inst1 := pkg.Instructions[1]
	if inst1.Op != "report" {
		t.Errorf("expected op report, got %s", inst1.Op)
	}
	if cpuRef, ok := inst1.Args["cpu"]; !ok || cpuRef != "cpu" {
		t.Errorf("expected report arg cpu=cpu, got %v", inst1.Args)
	}
}

func TestGenerate_LetAssignments(t *testing.T) {
	source := `task "test" on "host1" {
		let x = 42
		let name = "hello"
		let flag = true
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(pkg.Instructions) != 3 {
		t.Fatalf("expected 3 instructions, got %d", len(pkg.Instructions))
	}

	// let x = 42
	inst0 := pkg.Instructions[0]
	if inst0.Op != "set" {
		t.Errorf("inst0: expected op set, got %s", inst0.Op)
	}
	if inst0.Assign != "x" {
		t.Errorf("inst0: expected assign x, got %s", inst0.Assign)
	}
	if val, ok := inst0.Args["value"]; !ok || val != int64(42) {
		t.Errorf("inst0: expected value=42, got %v", val)
	}

	// let name = "hello"
	inst1 := pkg.Instructions[1]
	if inst1.Assign != "name" {
		t.Errorf("inst1: expected assign name, got %s", inst1.Assign)
	}
	if val, ok := inst1.Args["value"]; !ok || val != "hello" {
		t.Errorf("inst1: expected value=hello, got %v", val)
	}

	// let flag = true
	inst2 := pkg.Instructions[2]
	if inst2.Assign != "flag" {
		t.Errorf("inst2: expected assign flag, got %s", inst2.Assign)
	}
	if val, ok := inst2.Args["value"]; !ok || val != true {
		t.Errorf("inst2: expected value=true, got %v", val)
	}
}

func TestGenerate_FileOperations(t *testing.T) {
	source := `task "files" on "host1" {
		let content = file.read("/etc/hosts")
		file.exists("/tmp/test")
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(pkg.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(pkg.Instructions))
	}

	// file.read with assign
	inst0 := pkg.Instructions[0]
	if inst0.Op != "file.read" {
		t.Errorf("expected op file.read, got %s", inst0.Op)
	}
	if inst0.Assign != "content" {
		t.Errorf("expected assign content, got %s", inst0.Assign)
	}
	if path, ok := inst0.Args["path"]; !ok || path != "/etc/hosts" {
		t.Errorf("expected path=/etc/hosts, got %v", path)
	}

	// file.exists standalone
	inst1 := pkg.Instructions[1]
	if inst1.Op != "file.exists" {
		t.Errorf("expected op file.exists, got %s", inst1.Op)
	}
	if inst1.Assign != "" {
		t.Errorf("expected no assign, got %s", inst1.Assign)
	}
	if path, ok := inst1.Args["path"]; !ok || path != "/tmp/test" {
		t.Errorf("expected path=/tmp/test, got %v", path)
	}
}

func TestGenerate_DryRunFlag(t *testing.T) {
	source := `task "test" on "host1" {
		sys.cpu.usage()
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

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
	source := `task "test" on "host1" {
		report { host: sys.hostname(), cpu: sys.cpu.usage() }
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should generate: temp call for sys.hostname(), temp call for sys.cpu.usage(), then report.
	if len(pkg.Instructions) != 3 {
		t.Fatalf("expected 3 instructions, got %d", len(pkg.Instructions))
	}

	// First: sys.hostname -> __tmp_0
	inst0 := pkg.Instructions[0]
	if inst0.Op != "sys.hostname" {
		t.Errorf("inst0: expected op sys.hostname, got %s", inst0.Op)
	}
	if inst0.Assign != "__tmp_0" {
		t.Errorf("inst0: expected assign __tmp_0, got %s", inst0.Assign)
	}

	// Second: sys.cpu.usage -> __tmp_1
	inst1 := pkg.Instructions[1]
	if inst1.Op != "sys.cpu.usage" {
		t.Errorf("inst1: expected op sys.cpu.usage, got %s", inst1.Op)
	}
	if inst1.Assign != "__tmp_1" {
		t.Errorf("inst1: expected assign __tmp_1, got %s", inst1.Assign)
	}

	// Third: report referencing temps
	inst2 := pkg.Instructions[2]
	if inst2.Op != "report" {
		t.Errorf("inst2: expected op report, got %s", inst2.Op)
	}
	if hostRef, ok := inst2.Args["host"]; !ok || hostRef != "__tmp_0" {
		t.Errorf("inst2: expected host=__tmp_0, got %v", inst2.Args["host"])
	}
	if cpuRef, ok := inst2.Args["cpu"]; !ok || cpuRef != "__tmp_1" {
		t.Errorf("inst2: expected cpu=__tmp_1, got %v", inst2.Args["cpu"])
	}
}

func TestGenerate_AlertStatement(t *testing.T) {
	source := `task "test" on "host1" {
		alert("something went wrong")
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(pkg.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(pkg.Instructions))
	}

	inst := pkg.Instructions[0]
	if inst.Op != "alert" {
		t.Errorf("expected op alert, got %s", inst.Op)
	}
	if msg, ok := inst.Args["message"]; !ok || msg != "something went wrong" {
		t.Errorf("expected message=something went wrong, got %v", msg)
	}
}

func TestGenerate_PrintBuiltin(t *testing.T) {
	source := `task "test" on "host1" {
		print("hello world")
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(pkg.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(pkg.Instructions))
	}

	inst := pkg.Instructions[0]
	if inst.Op != "log" {
		t.Errorf("expected op log, got %s", inst.Op)
	}
	if msg, ok := inst.Args["message"]; !ok || msg != "hello world" {
		t.Errorf("expected message=hello world, got %v", msg)
	}
}

func TestGenerate_ForLoopWarning(t *testing.T) {
	source := `task "test" on "host1" {
		for let i = 0; i < 10; i = i + 1 {
			print("loop")
		}
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should produce a warning instruction
	if len(pkg.Instructions) == 0 {
		t.Fatal("expected at least one instruction (warning)")
	}

	found := false
	for _, inst := range pkg.Instructions {
		if inst.Op == "log" {
			if msg, ok := inst.Args["message"].(string); ok {
				if contains(msg, "for loops") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected a warning about for loops")
	}
}

func TestGenerateFromStatements(t *testing.T) {
	source := `task "test" on "host1" {
		let x = sys.cpu.usage()
		report { x: x }
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.GenerateFromStatements(task.Body.Statements, false)
	if err != nil {
		t.Fatalf("GenerateFromStatements failed: %v", err)
	}

	if len(pkg.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(pkg.Instructions))
	}
}

func TestGenerate_EmptyTask(t *testing.T) {
	source := `task "test" on "host1" {
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(pkg.Instructions) != 0 {
		t.Errorf("expected 0 instructions, got %d", len(pkg.Instructions))
	}
}

func TestGenerate_TaskIDUniqueness(t *testing.T) {
	source := `task "test" on "host1" {
		sys.cpu.usage()
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen1 := &InstructionGenerator{}
	pkg1, _ := gen1.Generate(task, false)

	gen2 := &InstructionGenerator{}
	pkg2, _ := gen2.Generate(task, false)

	if pkg1.TaskID == pkg2.TaskID {
		t.Error("expected unique task IDs across generations")
	}
}

func TestGenerate_CopyFileOperation(t *testing.T) {
	source := `task "test" on "host1" {
		file.copy("/src/file", "/dst/file")
	}`
	prog := mustParse(t, source)
	task := findTask(t, prog)

	gen := &InstructionGenerator{}
	pkg, err := gen.Generate(task, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(pkg.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(pkg.Instructions))
	}

	inst := pkg.Instructions[0]
	if inst.Op != "file.copy" {
		t.Errorf("expected op file.copy, got %s", inst.Op)
	}
	if src, ok := inst.Args["src"]; !ok || src != "/src/file" {
		t.Errorf("expected src=/src/file, got %v", src)
	}
	if dst, ok := inst.Args["dst"]; !ok || dst != "/dst/file" {
		t.Errorf("expected dst=/dst/file, got %v", dst)
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
