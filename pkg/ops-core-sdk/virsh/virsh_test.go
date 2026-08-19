package virsh

import "testing"

func TestStartEmpty(t *testing.T) {
	r := Start("")
	if r.Error == "" {
		t.Error("expected error for empty domain")
	}
}

func TestStopEmpty(t *testing.T) {
	r := Stop("")
	if r.Error == "" {
		t.Error("expected error for empty domain")
	}
}

func TestRebootEmpty(t *testing.T) {
	r := Reboot("")
	if r.Error == "" {
		t.Error("expected error for empty domain")
	}
}

func TestShutdownEmpty(t *testing.T) {
	r := Shutdown("")
	if r.Error == "" {
		t.Error("expected error for empty domain")
	}
}

func TestSuspendEmpty(t *testing.T) {
	r := Suspend("")
	if r.Error == "" {
		t.Error("expected error for empty domain")
	}
}

func TestResumeEmpty(t *testing.T) {
	r := Resume("")
	if r.Error == "" {
		t.Error("expected error for empty domain")
	}
}

func TestInfoEmpty(t *testing.T) {
	_, err := Info("")
	if err == nil {
		t.Error("expected error for empty domain")
	}
}

func TestList(t *testing.T) {
	r := List()
	// May fail if virsh not installed
	_ = r.Count
}

func TestVersion(t *testing.T) {
	_, err := Version()
	// May fail if virsh not installed
	_ = err
}
