package nvme

import "testing"

func TestSmartLogEmpty(t *testing.T) {
	_, err := SmartLog("")
	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestFirmwareLogEmpty(t *testing.T) {
	_, err := FirmwareLog("")
	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestErrorLogEmpty(t *testing.T) {
	_, err := ErrorLog("")
	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestList(t *testing.T) {
	r := List()
	// May fail if nvme not installed
	_ = r.Count
}

func TestVersion(t *testing.T) {
	_, err := Version()
	_ = err
}
