package supervisor

import "testing"

func TestStartMissingName(t *testing.T) {
	r := Start("")
	if r.Success {
		t.Error("expected failure for empty name")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestStopMissingName(t *testing.T) {
	r := Stop("")
	if r.Success {
		t.Error("expected failure for empty name")
	}
}

func TestRestartMissingName(t *testing.T) {
	r := Restart("")
	if r.Success {
		t.Error("expected failure for empty name")
	}
}

func TestClearLogMissingName(t *testing.T) {
	r := ClearLog("")
	if r.Success {
		t.Error("expected failure for empty name")
	}
}

func TestStatus(t *testing.T) {
	r := Status()
	// May fail if supervisor not installed
	_ = r.Count
}

func TestReload(t *testing.T) {
	r := Reload()
	// May fail if supervisor not installed
	_ = r
}

func TestReread(t *testing.T) {
	r := Reread()
	// May fail if supervisor not installed
	_ = r
}

func TestUpdate(t *testing.T) {
	r := Update("")
	// May fail if supervisor not installed
	_ = r
}
