package expect

import (
	"strings"
	"testing"
)

func TestRunEmpty(t *testing.T) {
	r := Run("", nil, 5)
	if r.Status != "failed" {
		t.Error("expected failure for empty command")
	}
}

func TestRunSimpleCommand(t *testing.T) {
	r := Run("echo hello", nil, 5)
	if r.Status != "success" {
		t.Errorf("expected success: %s", r.Error)
	}
	if !strings.Contains(r.Stdout, "hello") {
		t.Errorf("expected 'hello', got '%s'", r.Stdout)
	}
}

func TestRunWithResponse(t *testing.T) {
	// cat reads stdin and echoes it
	r := Run("cat", map[string]string{"": "test input"}, 5)
	if r.Status != "success" {
		t.Errorf("expected success: %s", r.Error)
	}
	if !strings.Contains(r.Stdout, "test input") {
		t.Errorf("expected 'test input', got '%s'", r.Stdout)
	}
}

func TestRunSimple(t *testing.T) {
	r := RunSimple("cat", "", "hello world", 5)
	if r.Status != "success" {
		t.Errorf("expected success: %s", r.Error)
	}
	if !strings.Contains(r.Stdout, "hello world") {
		t.Errorf("expected 'hello world', got '%s'", r.Stdout)
	}
}

func TestRunTimeout(t *testing.T) {
	// sleep should timeout
	r := Run("sleep 10", nil, 1)
	if r.Status != "failed" || !r.TimedOut {
		t.Error("expected timeout")
	}
}
