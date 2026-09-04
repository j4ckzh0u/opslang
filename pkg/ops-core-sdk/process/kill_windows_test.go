//go:build windows

package process

import "testing"

func TestWindowsSupportedSignals(t *testing.T) {
	for _, name := range []string{"TERM", "KILL", "INT"} {
		if !supportedSignal(name) {
			t.Errorf("supportedSignal(%q) = false, want true", name)
		}
	}
}

func TestWindowsUnknownSignal(t *testing.T) {
	if supportedSignal("") {
		t.Error("supportedSignal(empty) = true, want false")
	}
}
