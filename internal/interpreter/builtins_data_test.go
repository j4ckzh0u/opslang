package interpreter

import (
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
)

// evalVars is a helper: build a program from let statements, execute it,
// and return the variable environment. Each pair is {name, expression}.
func evalVars(t *testing.T, pairs ...[2]interface{}) map[string]interface{} {
	t.Helper()
	var stmts []ast.Statement
	for _, pair := range pairs {
		name, nameOK := pair[0].(string)
		expr, exprOK := pair[1].(ast.Expression)
		if !nameOK || !exprOK {
			t.Fatal("evalVars pairs must be {string, ast.Expression}")
		}
		stmts = append(stmts, let(name, expr))
	}
	r, err := newInterp().Execute(&ast.Program{Position: pos(), Statements: stmts})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return r.Variables
}

func strList(items ...string) []interface{} {
	out := make([]interface{}, len(items))
	for i, s := range items {
		out[i] = s
	}
	return out
}

func intList(items ...int64) []interface{} {
	out := make([]interface{}, len(items))
	for i, n := range items {
		out[i] = n
	}
	return out
}

func TestBuiltinSplitJoin(t *testing.T) {
	vars := evalVars(t,
		[2]interface{}{"parts", call(ident("split"), strLit("a,b,c"), strLit(","))},
		[2]interface{}{"joined", call(ident("join"), ident("parts"), strLit("-"))},
		[2]interface{}{"emptySep", call(ident("split"), strLit("xy"), strLit(""))},
	)
	parts, ok := vars["parts"].([]interface{})
	if !ok || len(parts) != 3 || parts[0] != "a" || parts[2] != "c" {
		t.Fatalf("split(a,b,c) = %v", vars["parts"])
	}
	if vars["joined"] != "a-b-c" {
		t.Fatalf("join(parts, -) = %v", vars["joined"])
	}
	empty, ok := vars["emptySep"].([]interface{})
	if !ok || len(empty) != 2 || empty[0] != "x" || empty[1] != "y" {
		t.Fatalf("split with empty separator = %v", vars["emptySep"])
	}
}

func TestBuiltinStringMutators(t *testing.T) {
	vars := evalVars(t,
		[2]interface{}{"r1", call(ident("replace"), strLit("10.0.0.1"), strLit("."), strLit("_"))},
		[2]interface{}{"r2", call(ident("upper"), strLit("web-01"))},
		[2]interface{}{"r3", call(ident("lower"), strLit("WEB-01"))},
		[2]interface{}{"r4", call(ident("trim"), strLit("  prod  \n"))},
	)
	want := map[string]string{
		"r1": "10_0_0_1",
		"r2": "WEB-01",
		"r3": "web-01",
		"r4": "prod",
	}
	for k, w := range want {
		if vars[k] != w {
			t.Errorf("%s = %v, want %q", k, vars[k], w)
		}
	}
}

func TestBuiltinContainsAndIndexOf(t *testing.T) {
	vars := evalVars(t,
		[2]interface{}{"inStr", call(ident("contains"), strLit("nginx.conf"), strLit("conf"))},
		[2]interface{}{"notInStr", call(ident("contains"), strLit("nginx"), strLit("apache"))},
		[2]interface{}{"inList", call(ident("contains"),
			&ast.ListLiteral{Position: pos(), Elements: []ast.Expression{intLit(1), intLit(2)}}, intLit(2))},
		[2]interface{}{"idxStr", call(ident("index_of"), strLit("hello world"), strLit("world"))},
		[2]interface{}{"idxMissing", call(ident("index_of"), strLit("hello"), strLit("zzz"))},
		[2]interface{}{"idxList", call(ident("index_of"),
			&ast.ListLiteral{Position: pos(), Elements: []ast.Expression{strLit("a"), strLit("b")}}, strLit("b"))},
	)
	checks := map[string]interface{}{
		"inStr": true, "notInStr": false, "inList": true,
		"idxStr": int64(6), "idxMissing": int64(-1), "idxList": int64(1),
	}
	for k, w := range checks {
		if vars[k] != w {
			t.Errorf("%s = %v (%T), want %v", k, vars[k], vars[k], w)
		}
	}
}

func TestBuiltinSortReverseImmutableInput(t *testing.T) {
	original := &ast.ListLiteral{Position: pos(), Elements: []ast.Expression{intLit(3), intLit(1), intLit(2)}}
	p := &ast.Program{Position: pos(), Statements: []ast.Statement{
		let("orig", original),
		let("sorted", call(ident("sort"), ident("orig"))),
		let("rev", call(ident("reverse"), ident("sorted"))),
	}}
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	gotSorted := r.Variables["sorted"].([]interface{})
	wantSorted := intList(1, 2, 3)
	for i := range wantSorted {
		if gotSorted[i] != wantSorted[i] {
			t.Fatalf("sort([3,1,2]) = %v", gotSorted)
		}
	}
	gotRev := r.Variables["rev"].([]interface{})
	wantRev := intList(3, 2, 1)
	for i := range wantRev {
		if gotRev[i] != wantRev[i] {
			t.Fatalf("reverse(sorted) = %v", gotRev)
		}
	}
	// The input list must be untouched (functional semantics).
	gotOrig := r.Variables["orig"].([]interface{})
	wantOrig := intList(3, 1, 2)
	for i := range wantOrig {
		if gotOrig[i] != wantOrig[i] {
			t.Fatalf("sort/reverse mutated their input: %v", gotOrig)
		}
	}
}

func TestBuiltinSortMixedTypesIsError(t *testing.T) {
	mixed := &ast.ListLiteral{Position: pos(), Elements: []ast.Expression{
		intLit(1), strLit("a"),
	}}
	p := prog(let("bad", call(ident("sort"), mixed)))
	if _, err := newInterp().Execute(p); err == nil {
		t.Fatal("sorting a mixed-type list must be an explicit error")
	}
}

func TestBuiltinKeysValuesSortedAndPaired(t *testing.T) {
	dict := &ast.DictLiteral{
		Position: pos(),
		Keys:     []ast.Expression{strLit("zeta"), strLit("alpha"), strLit("mid")},
		Values:   []ast.Expression{intLit(26), intLit(1), intLit(13)},
	}
	p := prog(
		let("k", call(ident("keys"), dict)),
		let("v", call(ident("values"), dict)),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	keys := r.Variables["k"].([]interface{})
	wantK := strList("alpha", "mid", "zeta")
	for i := range wantK {
		if keys[i] != wantK[i] {
			t.Fatalf("keys() = %v, want %v (must be sorted)", keys, wantK)
		}
	}
	vals := r.Variables["v"].([]interface{})
	wantV := intList(1, 13, 26)
	for i := range wantV {
		if vals[i] != wantV[i] {
			t.Fatalf("values() = %v, want %v (order must match sorted keys)", vals, wantV)
		}
	}
}

func TestDataBuiltinsRejectWrongArgCounts(t *testing.T) {
	cases := map[string]*ast.CallExpression{
		"split/1arg":  call(ident("split"), strLit("a")),
		"join/1arg":   call(ident("join"), strLit("a")),
		"upper/0arg":  call(ident("upper")),
		"contains/1":  call(ident("contains"), strLit("a")),
		"index_of/3":  call(ident("index_of"), strLit("a"), strLit("b"), strLit("c")),
		"keys/nonmap": call(ident("keys"), strLit("not-a-dict")),
	}
	for name, c := range cases {
		p := prog(let("x", c))
		if _, err := newInterp().Execute(p); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
