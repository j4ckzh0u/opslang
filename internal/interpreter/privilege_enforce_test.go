package interpreter

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/opslang/opslang/internal/parser"
)

// Privilege auto-enforcement: a script declaring read_only must not be able
// to call mutating builtins; admin/root may; undeclared scripts default to
// read_only (existing semantics, now actually enforced).

func TestPrivilegeReadOnlyRejectsMutatingCall(t *testing.T) {
	_, err := runSource(t, `privilege: read_only
let r = file.write("/tmp/opslang-priv-never.txt", "nope")`)
	if err == nil {
		t.Fatal("read_only script calling file.write must fail")
	}
	msg := err.Error()
	for _, want := range []string{"privilege denied", "file.write", "2:9"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q", msg, want)
		}
	}
}

func TestPrivilegeReadOnlyRejectsMutatingCallInsideBranch(t *testing.T) {
	// The call is only rejected when actually executed, but the position
	// must point at the call site.
	_, err := runSource(t, `privilege: read_only
if true {
	process.kill(99999, "TERM")
}`)
	if err == nil {
		t.Fatal("read_only script calling process.kill must fail")
	}
	if !strings.Contains(err.Error(), "process.kill") {
		t.Errorf("error must name the denied function: %v", err)
	}
}

func TestPrivilegeAdminAllowsMutatingCall(t *testing.T) {
	path := filepath.Join(t.TempDir(), "admin-write.txt")
	res := mustRun(t, `privilege: admin
let r = file.write("`+path+`", "written")
print(r.size)`)
	if len(res.Output) == 0 {
		t.Error("expected output")
	}
}

func TestPrivilegeRootAllowsMutatingCall(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "root-mkdir")
	mustRun(t, `privilege: root
file.mkdir("`+path+`")`)
}

func TestPrivilegeReadOnlyAllowsReadOnlyCalls(t *testing.T) {
	res := mustRun(t, `privilege: read_only
let h = sys.hostname()
let c = file.exists("/tmp")
print("ok")`)
	if len(res.Output) == 0 {
		t.Error("expected output")
	}
}

func TestPrivilegeDefaultIsReadOnlyWhenUndeclared(t *testing.T) {
	// No privilege statement: existing default is read_only, so mutating
	// calls are denied.
	_, err := runSource(t, `let r = file.write("/tmp/opslang-priv-undeclared.txt", "nope")`)
	if err == nil {
		t.Fatal("undeclared script defaults to read_only; file.write must fail")
	}
	if !strings.Contains(err.Error(), "privilege denied") {
		t.Errorf("unexpected error: %v", err)
	}
	// Read-only calls still work without a declaration.
	mustRun(t, `let h = sys.hostname()
print("ok")`)
}

func TestPrivilegeReadOnlyAllowsCustomBuiltins(t *testing.T) {
	// Host-injected builtins that are not OpsLang operations must not be
	// swept up by enforcement.
	p := parser.New(`print(double(21))`, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	interp := newInterp()
	RegisterSDKBuiltins(interp)
	interp.RegisterBuiltin("double", func(args ...interface{}) (interface{}, error) {
		return args[0].(int64) * 2, nil
	})
	if _, err := interp.Execute(prog); err != nil {
		t.Fatalf("custom builtin must not be privilege-checked: %v", err)
	}
}
