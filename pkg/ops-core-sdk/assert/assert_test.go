package assert

import "testing"

func TestAssertPass(t *testing.T) {
	r := Assert(true, "ok", "fail")
	if !r.Success || r.Failed {
		t.Error("expected success")
	}
	if r.Message != "ok" {
		t.Errorf("expected ok, got %s", r.Message)
	}
}

func TestAssertFail(t *testing.T) {
	r := Assert(false, "", "nope")
	if r.Success || !r.Failed {
		t.Error("expected failure")
	}
	if r.Message != "nope" {
		t.Errorf("expected nope, got %s", r.Message)
	}
}

func TestAssertEqual(t *testing.T) {
	r := AssertEqual(1, 1, "")
	if !r.Success {
		t.Error("1==1 should pass")
	}
	r = AssertEqual(1, 2, "1!=2")
	if r.Success {
		t.Error("1!=2 should fail")
	}
}
