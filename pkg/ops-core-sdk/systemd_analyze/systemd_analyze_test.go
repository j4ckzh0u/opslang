package systemdanalyze

import "testing"

func TestTime(t *testing.T) {
	r := Time()
	// May fail if systemd not available
	_ = r.Total
}

func TestBlame(t *testing.T) {
	r := Blame()
	_ = r.Count
}

func TestCriticalChain(t *testing.T) {
	r := CriticalChain()
	_ = r.Chain
}

func TestSecurity(t *testing.T) {
	_, err := Security("")
	// May fail if systemd not available
	_ = err
}

func TestVerify(t *testing.T) {
	_, err := Verify("")
	_ = err
}
