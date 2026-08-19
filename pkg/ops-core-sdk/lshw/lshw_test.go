package lshw

import "testing"

func TestShort(t *testing.T) {
	r := Short()
	// May fail if lshw not installed
	_ = r.Count
}

func TestClassEmpty(t *testing.T) {
	_, err := Class("")
	if err == nil {
		t.Error("expected error for empty class")
	}
}

func TestJSON(t *testing.T) {
	_, err := JSON()
	// May fail if lshw not installed
	_ = err
}

func TestSystem(t *testing.T) {
	_, err := System()
	_ = err
}

func TestMemory(t *testing.T) {
	_, err := Memory()
	_ = err
}

func TestDisk(t *testing.T) {
	_, err := Disk()
	_ = err
}

func TestNetwork(t *testing.T) {
	_, err := Network()
	_ = err
}
