package meta

import "testing"

func TestEndHost(t *testing.T) {
	r := EndHost()
	if r.Status != "success" || r.Action != "end_host" {
		t.Error("unexpected result")
	}
}

func TestEndPlay(t *testing.T) {
	r := EndPlay()
	if r.Status != "success" || r.Action != "end_play" {
		t.Error("unexpected result")
	}
}

func TestClearHostErrors(t *testing.T) {
	r := ClearHostErrors()
	if r.Status != "success" || r.Action != "clear_host_errors" {
		t.Error("unexpected result")
	}
}

func TestRefreshInventory(t *testing.T) {
	r := RefreshInventory()
	if r.Status != "success" {
		t.Error("unexpected result")
	}
}

func TestFlushHandlers(t *testing.T) {
	r := FlushHandlers()
	if r.Status != "success" {
		t.Error("unexpected result")
	}
}

func TestResetConnection(t *testing.T) {
	r := ResetConnection()
	if r.Status != "success" {
		t.Error("unexpected result")
	}
}

func TestNoop(t *testing.T) {
	r := Noop()
	if r.Status != "success" || r.Action != "noop" {
		t.Error("unexpected result")
	}
}

func TestFail(t *testing.T) {
	r := Fail("test error")
	if r.Status != "failed" || r.Error != "test error" {
		t.Error("unexpected result")
	}
}

func TestFailDefault(t *testing.T) {
	r := Fail("")
	if r.Status != "failed" || r.Error == "" {
		t.Error("expected error message")
	}
}

func TestAssertTrue(t *testing.T) {
	r := Assert(true, "should pass")
	if r.Status != "success" {
		t.Error("expected success for true condition")
	}
}

func TestAssertFalse(t *testing.T) {
	r := Assert(false, "should fail")
	if r.Status != "failed" {
		t.Error("expected failure for false condition")
	}
}

func TestDebug(t *testing.T) {
	r := Debug("test message", nil)
	if r.Status != "success" || r.Action != "debug" {
		t.Error("unexpected result")
	}
}
