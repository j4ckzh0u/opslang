package runner

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
)

// Runner-side privilege enforcement: the instruction generator refuses to
// emit mutating instructions for read_only scripts (deploy-time check), and
// the executor refuses to execute them when the package declares a
// privilege that contradicts them (second check on the host).

// ---------- generator (controller side) ----------

func TestGenerate_ReadOnlyDeniesMutatingCall(t *testing.T) {
	task := findTask(t, mustParse(t, `task "t" on "h" {
		file.write("/tmp/never.txt", "nope")
	}`))
	// Undeclared privilege defaults to read_only, like everywhere else.
	gen := &InstructionGenerator{}
	_, err := gen.Generate(task, false)
	if err == nil {
		t.Fatal("read_only task calling file.write must fail generation")
	}
	for _, want := range []string{"privilege denied", "file.write"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err.Error(), want)
		}
	}
}

func TestGenerate_ExplicitReadOnlyDeniesMutatingCall(t *testing.T) {
	_, err := (&InstructionGenerator{}).GenerateFromStatements(
		mustParse(t, `privilege: read_only
file.mkdir("/tmp/never")`).Statements, false)
	if err == nil {
		t.Fatal("declared read_only script calling file.mkdir must fail generation")
	}
}

func TestGenerate_AdminAllowsMutatingCall(t *testing.T) {
	pkg, err := (&InstructionGenerator{}).GenerateFromStatements(
		mustParse(t, `privilege: admin
file.mkdir("/tmp/ok")`).Statements, false)
	if err != nil {
		t.Fatalf("admin script calling file.mkdir must generate: %v", err)
	}
	if pkg.Privilege != "admin" {
		t.Errorf("package privilege = %q, want admin", pkg.Privilege)
	}
}

func TestGenerate_ReadOnlyAllowsReadCalls(t *testing.T) {
	pkg, err := (&InstructionGenerator{}).GenerateFromStatements(
		mustParse(t, `privilege: read_only
let h = sys.hostname()
report { host: h }`).Statements, false)
	if err != nil {
		t.Fatalf("read_only script with read-only calls must generate: %v", err)
	}
	if pkg.Privilege != "read_only" {
		t.Errorf("package privilege = %q, want read_only", pkg.Privilege)
	}
}

func TestGenerate_UndeclaredDefaultsToReadOnlyInPackage(t *testing.T) {
	pkg, err := (&InstructionGenerator{}).GenerateFromStatements(
		mustParse(t, `let h = sys.hostname()`).Statements, false)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Privilege != string(ast.PrivilegeReadOnly) {
		t.Errorf("package privilege = %q, want read_only default", pkg.Privilege)
	}
}

// ---------- executor (host side, second check) ----------

func TestRun_ReadOnlyPackageRejectsMutatingInstruction(t *testing.T) {
	pkg := &InstructionPackage{
		Version:   "1.0",
		Privilege: string(ast.PrivilegeReadOnly),
		Instructions: []Instruction{
			{Op: "sys.hostname", Assign: "h"},
			{Op: "file.write", Args: map[string]interface{}{"path": "/tmp/runner-priv-never.txt", "content": "x"}},
		},
	}
	out := Run(pkg, NewRegistry())
	if out.Status == "ok" {
		t.Fatal("read_only package with file.write instruction must not be ok")
	}
	found := false
	for _, e := range out.Errors {
		if strings.Contains(e, "privilege denied") && strings.Contains(e, "file.write") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected structured privilege error mentioning file.write, got: %v", out.Errors)
	}
}

func TestRun_ReadOnlyPackageAllowsReadInstructions(t *testing.T) {
	pkg := &InstructionPackage{
		Version:   "1.0",
		Privilege: string(ast.PrivilegeReadOnly),
		Instructions: []Instruction{
			{Op: "sys.hostname", Assign: "h"},
		},
	}
	out := Run(pkg, NewRegistry())
	if out.Status != "ok" {
		t.Fatalf("read_only package with read-only instructions must be ok: %v", out.Errors)
	}
}

func TestRun_AdminPackageAllowsMutatingInstruction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runner-admin.txt")
	pkg := &InstructionPackage{
		Version:   "1.0",
		Privilege: string(ast.PrivilegeAdmin),
		Instructions: []Instruction{
			{Op: "file.write", Args: map[string]interface{}{"path": path, "content": "x"}, Assign: "w"},
		},
	}
	out := Run(pkg, NewRegistry())
	if out.Status != "ok" {
		t.Fatalf("admin package with file.write must be ok: %v", out.Errors)
	}
}

func TestRun_LegacyPackageWithoutPrivilegeStillRuns(t *testing.T) {
	// Backward compatibility: packages predating the privilege field must
	// behave exactly as before (no runner-side check).
	path := filepath.Join(t.TempDir(), "runner-legacy.txt")
	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "file.write", Args: map[string]interface{}{"path": path, "content": "x"}},
		},
	}
	out := Run(pkg, NewRegistry())
	if out.Status != "ok" {
		t.Fatalf("legacy package without privilege field must keep working: %v", out.Errors)
	}
}

func TestRun_InvalidPrivilegeFailsStructured(t *testing.T) {
	pkg := &InstructionPackage{
		Version:   "1.0",
		Privilege: "superuser",
		Instructions: []Instruction{
			{Op: "sys.hostname", Assign: "h"},
		},
	}
	out := Run(pkg, NewRegistry())
	if out.Status != "failed" {
		t.Fatalf("invalid privilege must fail the run, got status %q", out.Status)
	}
	if len(out.Errors) == 0 || !strings.Contains(out.Errors[0], "invalid package privilege") {
		t.Errorf("expected structured invalid-privilege error, got: %v", out.Errors)
	}
}

func TestValidatePackage_RejectsInvalidPrivilege(t *testing.T) {
	pkg := &InstructionPackage{
		Version:      "1.0",
		Privilege:    "root-ish",
		Instructions: []Instruction{{Op: "sys.hostname"}},
	}
	if err := ValidatePackage(pkg); err == nil || !strings.Contains(err.Error(), "invalid privilege") {
		t.Errorf("expected invalid-privilege validation error, got: %v", err)
	}

	// Empty (legacy) and all declared levels must pass the field check.
	for _, p := range []string{"", "read_only", "admin", "root"} {
		pkg.Privilege = p
		if err := ValidatePackage(pkg); err != nil {
			t.Errorf("privilege %q must validate: %v", p, err)
		}
	}
}
