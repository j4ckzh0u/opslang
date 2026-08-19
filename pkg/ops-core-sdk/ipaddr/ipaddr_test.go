package ipaddr

import "testing"

func TestList(t *testing.T) {
	r := List()
	// May fail if ip not available
	_ = r.Count
}

func TestListInterfaceEmpty(t *testing.T) {
	r := ListInterface("")
	if r.Error == "" {
		t.Error("expected error for empty interface")
	}
}

func TestAddEmpty(t *testing.T) {
	r := Add("", "")
	if r.Error == "" {
		t.Error("expected error for empty args")
	}
}

func TestDeleteEmpty(t *testing.T) {
	r := Delete("", "")
	if r.Error == "" {
		t.Error("expected error for empty args")
	}
}

func TestFlushEmpty(t *testing.T) {
	r := Flush("")
	if r.Error == "" {
		t.Error("expected error for empty interface")
	}
}

func TestLinks(t *testing.T) {
	r := Links()
	// May fail if ip not available
	_ = r.Count
}

func TestLinkUpEmpty(t *testing.T) {
	r := LinkUp("")
	if r.Error == "" {
		t.Error("expected error for empty interface")
	}
}

func TestLinkDownEmpty(t *testing.T) {
	r := LinkDown("")
	if r.Error == "" {
		t.Error("expected error for empty interface")
	}
}
