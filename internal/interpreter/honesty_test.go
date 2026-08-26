package interpreter

import (
	"os"
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/parser"
)

// runSource executes source text and returns (result, error).
func runSource(t *testing.T, source string) (*Result, error) {
	t.Helper()
	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	interp := newInterp()
	RegisterSDKBuiltins(interp)
	return interp.Execute(prog)
}

func mustRun(t *testing.T, source string) *Result {
	t.Helper()
	res, err := runSource(t, source)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	return res
}

func mustFail(t *testing.T, source, wantInMsg string) {
	t.Helper()
	_, err := runSource(t, source)
	if err == nil {
		t.Fatalf("expected error containing %q, got success", wantInMsg)
	}
	if !strings.Contains(err.Error(), wantInMsg) {
		t.Fatalf("expected error containing %q, got: %v", wantInMsg, err)
	}
}

// ---------- strict numeric conversion ----------

func TestStrictIntConversion(t *testing.T) {
	mustRun(t, `let x = int("42") print(x)`)

	// Sscanf prefix-matching used to accept these silently.
	mustFail(t, `let x = int("42abc") print(x)`, `cannot convert "42abc" to int`)
	mustFail(t, `let x = int("3.14") print(x)`, `cannot convert "3.14" to int`)
	mustFail(t, `let x = float("1.5xyz") print(x)`, `cannot convert "1.5xyz" to float`)
}

// ---------- missing dict keys ----------

func TestMissingDictKeyIsError(t *testing.T) {
	// A typo in a field name must fail at the access site, not far away.
	mustFail(t, `let m = {"used": 1}
let v = m.used_prcnt
print(v)`, `map has no key "used_prcnt"`)

	mustFail(t, `let m = {"used": 1}
let v = m["missing"]
print(v)`, `map has no key "missing"`)
}

// ---------- user function arity ----------

func TestUserFunctionMissingArgumentIsError(t *testing.T) {
	mustFail(t, `fn add(a, b) { return a + b }
let s = add(1)
print(s)`, `missing argument "b"`)
}

// ---------- let redeclaration ----------

func TestLetRedeclarationIsError(t *testing.T) {
	mustFail(t, `let x = 1
let x = 2
print(x)`, `already declared`)
	mustRun(t, `let x = 1
x = 2
print(x)`)
}

// ---------- privilege statement ----------

func TestPrivilegeStatementAccepted(t *testing.T) {
	res := mustRun(t, `privilege: read_only
print("ok")`)
	if len(res.Output) == 0 {
		t.Error("expected output")
	}
}

// ---------- task with remote targets ----------

func TestTaskWithOnTargetsRequiresDeploy(t *testing.T) {
	mustFail(t, `task "check" on "host1" { print("x") }`, "opsctl deploy")
}

// ---------- ensure notify ----------

func TestEnsureNotifyFiresOnlyOnChange(t *testing.T) {
	res := mustRun(t, `let n = 0
fn notify() { let n2 = 1 return n2 }
ensure file.exists("/tmp").exists {
	file.mkdir("/tmp/definitely-exists-now-xyz")
}
print("done")`)
	_ = res // no notify clause: nothing extra to check

	// With a notify expression: evaluated only when the apply step ran and
	// verification passed. The body flips n so verify succeeds.
	res1 := mustRun(t, `let n = 0
ensure n == 1 {
	n = 1
} notify print("state changed")
print("done")`)
	foundNotified := false
	for _, o := range res1.Output {
		if o.Data == "state changed" {
			foundNotified = true
		}
	}
	if !foundNotified {
		t.Fatal("notify expression must fire after an applied change")
	}

	// Already-satisfied condition: apply and notify must both be skipped.
	res2 := mustRun(t, `let n = 1
ensure n == 1 {
	n = 2
} notify print("state changed")
print("done")`)
	for _, o := range res2.Output {
		if o.Data == "state changed" {
			t.Fatal("notify must NOT fire when nothing was applied")
		}
	}
	_ = res
}

// ---------- import go ----------

func TestImportGoRejected(t *testing.T) {
	// The go marker lives inside the import string: import "go <pkg>".
	mustFail(t, `import "go github.com/x/y"`, "not supported")
}

// ---------- strict equality ----------

func TestStrictEquality(t *testing.T) {
	mustRun(t, `
if 1 == 1.0 { print("numeric equal") }
if "x" == "x" { print("string equal") }
`)
	mustFail(t, `
if 1 == "1" {
	print("cross-type equal")
} else {
	missing_function_to_abort()
}`, "unknown")
}

// ---------- parallel isolation ----------

func TestParallelBarrierIsolation(t *testing.T) {
	// Assignments inside parallel are captured and merged deterministically.
	res := mustRun(t, `
let a = 0
parallel {
	let a = 1
}
print(str(a))
`)
	if len(res.Output) == 0 || res.Output[len(res.Output)-1].Data != "1" {
		t.Errorf("parallel merge failed: %+v", res.Output)
	}
}

// ---------- dry-run ----------

func TestDryRunSkipsEnsureApply(t *testing.T) {
	source := `ensure file.exists("/tmp/opslang-dryrun-never-created").exists {
	file.mkdir("/tmp/opslang-dryrun-never-created")
}`
	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	interp := newInterp()
	RegisterSDKBuiltins(interp)
	interp.SetDryRun(true)

	res, err := interp.Execute(prog)
	if err != nil {
		t.Fatalf("dry-run execution error: %v", err)
	}

	found := false
	for _, o := range res.Output {
		if o.Type == "dry-run" {
			found = true
		}
	}
	if !found {
		t.Fatal("dry-run must report the planned apply")
	}

	// The mutation must NOT have happened.
	if _, err := os.Stat("/tmp/opslang-dryrun-never-created"); err == nil {
		t.Fatal("dry-run executed the apply step - mutations must be suppressed")
	}
}
