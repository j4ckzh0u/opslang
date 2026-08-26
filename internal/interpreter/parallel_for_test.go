package interpreter

import (
	"strings"
	"testing"
)

// TestParallelForMergesInSourceOrder pins the deterministic-merge
// contract: later iterations overwrite earlier ones for the same
// assigned variable, regardless of goroutine scheduling.
func TestParallelForMergesInSourceOrder(t *testing.T) {
	src := `
let last = ""
parallel for n in [1, 2, 3] {
	last = str(n)
}
`
	r, err := runSource(t, src)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if r.Variables["last"] != "3" {
		t.Errorf("last = %v, want \"3\" (source-order merge)", r.Variables["last"])
	}
}

// TestParallelForLoopVarNotLeaked: the loop variable lives in the
// iteration env and does not leak into the outer scope.
func TestParallelForLoopVarNotLeaked(t *testing.T) {
	src := `
let seen = 0
parallel for x in [7, 8] {
	seen = x
}
`
	r, err := runSource(t, src)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if _, leaked := r.Variables["x"]; leaked {
		t.Error("loop variable x should not leak into outer scope")
	}
	if r.Variables["seen"] != 3.0 && r.Variables["seen"] != int64(8) && r.Variables["seen"] != float64(8) {
		t.Errorf("seen = %#v, want last iteration value 8", r.Variables["seen"])
	}
}

// TestParallelForRequiresList: iterating a non-list is an explicit error.
func TestParallelForRequiresList(t *testing.T) {
	src := `
parallel for x in 42 {
	print(x)
}
`
	_, err := runSource(t, src)
	if err == nil {
		t.Fatal("parallel for over non-list must error")
	}
	if !strings.Contains(err.Error(), "list") {
		t.Errorf("error should mention list requirement: %v", err)
	}
}

// TestParallelForEmptyListIsNoop covers the empty edge case.
func TestParallelForEmptyListIsNoop(t *testing.T) {
	src := `
let touched = false
parallel for x in [] {
	touched = true
}
`
	r, err := runSource(t, src)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if r.Variables["touched"] != false {
		t.Errorf("empty list must not run the body; touched = %v", r.Variables["touched"])
	}
}

// TestParallelForRejectsSharedDictMutation pins the honest-refusal
// contract: writing into a dict declared OUTSIDE the loop races across
// goroutines ("concurrent map writes"), so both engines reject the
// pattern with a clear message instead of crashing.
func TestParallelForRejectsSharedDictMutation(t *testing.T) {
	src := `
let results = {}
parallel for n in [1, 2, 3, 4, 5] {
	results[str(n)] = n * 10
}
`
	_, err := runSource(t, src)
	if err == nil {
		t.Fatal("shared dict mutation inside parallel for must be rejected")
	}
	if !strings.Contains(err.Error(), "shared state") {
		t.Errorf("error should explain the race: %v", err)
	}
}
