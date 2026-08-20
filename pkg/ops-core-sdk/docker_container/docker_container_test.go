package docker_container

import "testing"

func TestStartValidation(t *testing.T) {
	_, err := Start("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStopValidation(t *testing.T) {
	_, err := Stop("", 0)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemoveValidation(t *testing.T) {
	_, err := Remove("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRestartValidation(t *testing.T) {
	_, err := Restart("", 0)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestPauseValidation(t *testing.T) {
	_, err := Pause("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUnpauseValidation(t *testing.T) {
	_, err := Unpause("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInspectValidation(t *testing.T) {
	_, err := Inspect("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestLogsValidation(t *testing.T) {
	_, err := Logs("", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestFindDocker(t *testing.T) {
	_ = findDocker()
}
