package compiler

import (
	"strings"
	"testing"
)

// Compile-time privilege enforcement: statically-visible mutating calls in
// read_only scripts must fail code generation with position and function
// name; admin scripts and read-only calls must pass.

func TestPrivilegeReadOnlyFailsCodeGen(t *testing.T) {
	_, err := GenerateCode(`privilege: read_only
file.write("/tmp/x", "no")`, "test.ops")
	if err == nil {
		t.Fatal("read_only script with file.write must fail compilation")
	}
	msg := err.Error()
	for _, want := range []string{"privilege denied", "file.write", "2:1"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q", msg, want)
		}
	}
}

func TestPrivilegeViolationInsideNestedBodiesIsCaught(t *testing.T) {
	// The call sits in an ensure body inside a user function — the walk
	// must reach every nested construct.
	_, err := GenerateCode(`privilege: read_only
fn deploy() {
	ensure file.exists("/tmp/x").exists {
		file.mkdir("/tmp/x")
	}
}
deploy()`, "test.ops")
	if err == nil {
		t.Fatal("file.mkdir inside ensure body must fail compilation")
	}
	if !strings.Contains(err.Error(), "file.mkdir") {
		t.Errorf("error must name the denied function: %v", err)
	}
}

func TestPrivilegeViolationInsideCallArgumentsIsCaught(t *testing.T) {
	_, err := GenerateCode(`privilege: read_only
report { wrote: file.append("/tmp/x", "no") }`, "test.ops")
	if err == nil {
		t.Fatal("file.append inside report field must fail compilation")
	}
}

func TestPrivilegeAdminPassesCodeGen(t *testing.T) {
	if _, err := GenerateCode(`privilege: admin
file.write("/tmp/x", "yes")`, "test.ops"); err != nil {
		t.Fatalf("admin script with file.write must compile: %v", err)
	}
}

func TestPrivilegeRootPassesCodeGen(t *testing.T) {
	if _, err := GenerateCode(`privilege: root
process.kill(123, "TERM")`, "test.ops"); err != nil {
		t.Fatalf("root script with process.kill must compile: %v", err)
	}
}

func TestPrivilegeReadOnlyAllowsReadCalls(t *testing.T) {
	if _, err := GenerateCode(`privilege: read_only
let h = sys.hostname()
let e = file.exists("/tmp")
let t = file.template("/tmp/tpl", {"x": 1})`, "test.ops"); err != nil {
		t.Fatalf("read_only script with read-only calls must compile: %v", err)
	}
}

func TestPrivilegeUndeclaredDefaultsToReadOnly(t *testing.T) {
	// No privilege statement: existing default (read_only) applies.
	_, err := GenerateCode(`file.write("/tmp/x", "no")`, "test.ops")
	if err == nil {
		t.Fatal("undeclared script defaults to read_only; file.write must fail")
	}
	if _, err := GenerateCode(`let h = sys.hostname()`, "test.ops"); err != nil {
		t.Fatalf("read-only call in undeclared script must compile: %v", err)
	}
}
