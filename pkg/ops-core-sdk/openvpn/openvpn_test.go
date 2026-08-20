package openvpn

import (
	"testing"
)

func TestStatus(t *testing.T) {
	result, err := Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if result == nil {
		t.Fatal("Status() returned nil result")
	}
}

func TestGenKeyEmptyPath(t *testing.T) {
	_, err := GenKey("")
	if err == nil {
		t.Fatal("GenKey('') should return error for empty path")
	}
}

func TestGenTLSAuthEmptyPath(t *testing.T) {
	_, err := GenTLSAuth("")
	if err == nil {
		t.Fatal("GenTLSAuth('') should return error for empty path")
	}
}
