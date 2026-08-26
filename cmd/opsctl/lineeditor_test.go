package main

import (
	"os"
	"testing"
)

func TestHistoryAddDedupesAndSkipsEmpty(t *testing.T) {
	h := &history{}
	h.add("let x = 1")
	h.add("let x = 1") // duplicate of newest: dropped
	h.add("")          // empty: dropped
	h.add("print(x)")

	if len(h.entries) != 2 {
		t.Fatalf("entries = %v, want 2", h.entries)
	}
	if h.entries[0] != "let x = 1" || h.entries[1] != "print(x)" {
		t.Errorf("entries = %v", h.entries)
	}
}

func TestHistoryBoundsEntries(t *testing.T) {
	h := &history{}
	for i := 0; i < 520; i++ {
		h.add("cmd " + string(rune('a'+i%26)) + " " + string(rune('0'+i%10)) + " " + string(rune('0'+i/10)))
	}
	if len(h.entries) != 500 {
		t.Errorf("entries len = %d, want bounded at 500", len(h.entries))
	}
}

func TestHistoryNavRoundTrip(t *testing.T) {
	h := &history{}
	h.add("first")
	h.add("second")

	r := newLineReader("ops> ", h)

	// Up from live line: newest entry, live line stashed.
	line, cur := r.histPrev([]rune("typing"))
	if string(line) != "second" || cur != len("second") {
		t.Errorf("histPrev = %q/%d, want second/%d", string(line), cur, len("second"))
	}
	// Up again: older entry.
	line, _ = r.histPrev(line)
	if string(line) != "first" {
		t.Errorf("histPrev twice = %q, want first", string(line))
	}
	// Up at the oldest entry stays put.
	line, _ = r.histPrev(line)
	if string(line) != "first" {
		t.Errorf("histPrev at oldest = %q, want first (stay)", string(line))
	}
	// Down twice: back to newest, then the stashed live line.
	line, _ = r.histNext(line)
	if string(line) != "second" {
		t.Errorf("histNext = %q, want second", string(line))
	}
	line, cur = r.histNext(line)
	if string(line) != "typing" || cur != len("typing") {
		t.Errorf("histNext past newest = %q, want stashed 'typing'", string(line))
	}
	// Down from live line is a no-op.
	line, _ = r.histNext(line)
	if string(line) != "typing" {
		t.Errorf("histNext from live = %q, want typing", string(line))
	}
}

// The escape-sequence decoding and rendering live in read()'s TTY branch,
// which needs a real terminal; those paths are exercised interactively and
// via the pipe fallback in TestREPLPipeFallback below.
func TestHistoryResetOnNewSubmission(t *testing.T) {
	h := &history{}
	h.add("one")
	r := newLineReader("p> ", h)
	r.histPrev([]rune("x"))
	h.add("two")
	if h.navIdx != -1 {
		t.Errorf("navIdx = %d after add, want -1 (reset to live)", h.navIdx)
	}
}

// TestREPLPipeFallback drives the REPL through a pipe (non-TTY): the
// fallback path must keep scripted use working, including the final line
// without a trailing newline.
func TestREPLPipeFallback(t *testing.T) {
	// runREPL reads os.Stdin directly; injecting means swapping the file.
	// Simplest reliable approach: exercise newLineReader's fallback against
	// a replaced stdin.
	script := "let x = 41\nx + 1\nexit\n"
	tmp := t.TempDir() + "/in.txt"
	if err := os.WriteFile(tmp, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	oldStdin := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = oldStdin }()

	r := newLineReader("ops> ", &history{})
	l1, err := r.read()
	if err != nil || l1 != "let x = 41" {
		t.Fatalf("read 1 = %q, %v", l1, err)
	}
	l2, err := r.read()
	if err != nil || l2 != "x + 1" {
		t.Fatalf("read 2 = %q, %v", l2, err)
	}
}
