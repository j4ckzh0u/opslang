package interpreter

import "testing"

// lastReport returns the Data of the first report output entry, failing
// the test if no report is present.
func lastReport(t *testing.T, r *Result) map[string]interface{} {
	t.Helper()
	for _, e := range r.Output {
		if e.Type == "report" {
			if d, ok := e.Data.(map[string]interface{}); ok {
				return d
			}
			t.Fatalf("report data is not a map: %T", e.Data)
		}
	}
	t.Fatalf("no report output in result")
	return nil
}

// TestDictLiteralIdentifierKey pins the fix: dict literal keys that look
// like identifiers (e.g. { a: 1, b: 2 }) must be treated as the key name,
// not looked up as a variable. The earlier eval path tried to evaluate
// the key as an expression and rejected every literal identifier key.
func TestDictLiteralIdentifierKey(t *testing.T) {
	r, err := runSource(t, `
let x = { a: 1, b: 2, c: 3 }
report { a: x.a, b: x.b, c: x.c }
`)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	d := lastReport(t, r)
	got, _ := d["a"].(int64)
	if got != 1 {
		t.Errorf("dict a = %v (type %T), want 1", d["a"], d["a"])
	}
}

// TestDictLiteralStringKey: string-literal keys must keep working.
func TestDictLiteralStringKey(t *testing.T) {
	r, err := runSource(t, `
let x = { "hello": 1, "world": 2 }
report { h: x["hello"], w: x["world"] }
`)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	d := lastReport(t, r)
	if h, _ := d["h"].(int64); h != 1 {
		t.Errorf("h = %v, want 1", d["h"])
	}
	if w, _ := d["w"].(int64); w != 2 {
		t.Errorf("w = %v, want 2", d["w"])
	}
}

// TestDictLiteralExpressionValue covers non-literal values - the value
// side still has to be evaluated, even when the key is a literal.
func TestDictLiteralExpressionValue(t *testing.T) {
	r, err := runSource(t, `
let a = 5
let b = 10
let x = { sum: a + b, prod: a * b }
report { sum: x.sum, prod: x.prod }
`)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	d := lastReport(t, r)
	if s, _ := d["sum"].(int64); s != 15 {
		t.Errorf("sum = %v (type %T), want 15", d["sum"], d["sum"])
	}
	if p, _ := d["prod"].(int64); p != 50 {
		t.Errorf("prod = %v (type %T), want 50", d["prod"], d["prod"])
	}
}

// TestDictLiteralNested: nested dict literals (the original failure
// shape in the Linux e2e demo).
func TestDictLiteralNested(t *testing.T) {
	r, err := runSource(t, `
let x = { outer: { inner: 42 } }
report { v: x.outer.inner }
`)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	d := lastReport(t, r)
	if v, _ := d["v"].(int64); v != 42 {
		t.Errorf("nested access = %v (type %T), want 42", d["v"], d["v"])
	}
}
