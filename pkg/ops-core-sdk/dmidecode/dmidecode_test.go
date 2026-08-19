package dmidecode

import "testing"

func TestKeywordEmpty(t *testing.T) {
	_, err := Keyword("")
	if err == nil {
		t.Error("expected error for empty keyword")
	}
}

func TestSystem(t *testing.T) {
	r := System()
	// May fail if dmidecode not available or no permissions
	_ = r.Manufacturer
}

func TestBIOS(t *testing.T) {
	r := BIOS()
	_ = r.Vendor
}

func TestChassis(t *testing.T) {
	r := Chassis()
	_ = r.Manufacturer
}

func TestProcessor(t *testing.T) {
	r := Processor()
	_ = r.Manufacturer
}
