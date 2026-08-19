package supervisor

import (
	"testing"
)

func TestStartEmptyName(t *testing.T) {
	r := Start("")
	if r.Status != "failed" {
		t.Errorf("expected status=failed for empty name, got %s", r.Status)
	}
}

func TestStopEmptyName(t *testing.T) {
	r := Stop("")
	if r.Status != "failed" {
		t.Errorf("expected status=failed for empty name, got %s", r.Status)
	}
}

func TestRestartEmptyName(t *testing.T) {
	r := Restart("")
	if r.Status != "failed" {
		t.Errorf("expected status=failed for empty name, got %s", r.Status)
	}
}

func TestClearLogEmptyName(t *testing.T) {
	r := ClearLog("")
	if r.Status != "failed" {
		t.Errorf("expected status=failed for empty name, got %s", r.Status)
	}
}

func TestUpdateEmptyName(t *testing.T) {
	r := Update("")
	if r.Status != "failed" {
		t.Errorf("expected status=failed for empty name, got %s", r.Status)
	}
}

func TestStatus(t *testing.T) {
	// Status will succeed or fail based on supervisor availability
	r := Status()
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestReload(t *testing.T) {
	r := Reload()
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestReread(t *testing.T) {
	r := Reread()
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}
