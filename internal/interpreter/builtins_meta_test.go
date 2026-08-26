package interpreter

import (
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
)

// evalOne evaluates a single call expression in a fresh interpreter and
// returns the value bound to "x".
func evalOne(t *testing.T, expr ast.Expression) interface{} {
	t.Helper()
	r, err := newInterp().Execute(&ast.Program{
		Position:   pos(),
		Statements: []ast.Statement{let("x", expr)},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return r.Variables["x"]
}

// evalOneErr evaluates a single call expression expecting an error.
func evalOneErr(t *testing.T, expr ast.Expression) error {
	t.Helper()
	_, err := newInterp().Execute(&ast.Program{
		Position:   pos(),
		Statements: []ast.Statement{let("x", expr)},
	})
	return err
}

// TestDocBuiltin pins the introspection contract for a known op.
func TestDocBuiltin(t *testing.T) {
	result := evalOne(t, call(ident("doc"), strLit("sys.disk.usage")))
	dict, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("doc() returned %T, want dict", result)
	}
	if dict["name"] != "sys.disk.usage" {
		t.Errorf("name = %v", dict["name"])
	}
	args, ok := dict["args"].([]interface{})
	if !ok || len(args) != 1 || args[0] != "path" {
		t.Errorf("args = %#v, want [path]", dict["args"])
	}
	if dict["mutating"] != false {
		t.Errorf("mutating = %v, want false", dict["mutating"])
	}
}

// TestDocMutatingOp verifies the mutating flag surfaces for a known
// state-changing operation.
func TestDocMutatingOp(t *testing.T) {
	result := evalOne(t, call(ident("doc"), strLit("apt.install")))
	dict := result.(map[string]interface{})
	if dict["mutating"] != true {
		t.Errorf("apt.install mutating = %v, want true", dict["mutating"])
	}
}

// TestDocResolvesAlias ensures historical aliases document the canonical op.
func TestDocResolvesAlias(t *testing.T) {
	result := evalOne(t, call(ident("doc"), strLit("net.tcp.ping")))
	dict := result.(map[string]interface{})
	if dict["name"] != "net.tcp_check" {
		t.Errorf("alias doc name = %v, want net.tcp_check", dict["name"])
	}
}

// TestDocUnknownOpIsExplicitError: no silent empty results.
func TestDocUnknownOpIsExplicitError(t *testing.T) {
	err := evalOneErr(t, call(ident("doc"), strLit("no.such.op")))
	if err == nil {
		t.Fatal("doc() of unknown op must error")
	}
	if !strings.Contains(err.Error(), "ops()") {
		t.Errorf("error should point at ops(): %v", err)
	}
}

// TestOpsPrefixListing checks discovery: full list is sorted and prefix
// filtering narrows it correctly.
func TestOpsPrefixListing(t *testing.T) {
	all, ok := evalOne(t, call(ident("ops"))).([]interface{})
	if !ok {
		t.Fatalf("ops() returned unexpected type")
	}
	if len(all) < 100 {
		t.Errorf("ops() returned %d entries, expected the full table (1000+)", len(all))
	}
	for i := 1; i < len(all); i++ {
		prev, _ := all[i-1].(string)
		cur, _ := all[i].(string)
		if prev > cur {
			t.Fatalf("ops() not sorted at index %d: %q > %q", i, prev, cur)
		}
	}

	sysOps, ok := evalOne(t, call(ident("ops"), strLit("sys.disk."))).([]interface{})
	if !ok {
		t.Fatalf("ops(prefix) returned unexpected type")
	}
	if len(sysOps) < 2 {
		t.Errorf(`ops("sys.disk.") returned %d entries, want >= 2`, len(sysOps))
	}
	for _, v := range sysOps {
		s, _ := v.(string)
		if !strings.HasPrefix(s, "sys.disk.") {
			t.Errorf("prefix filter leaked: %q", s)
		}
	}
}
