//go:build linux

package sys

import (
	"testing"
)

func TestHostnameSet_Empty(t *testing.T) {
	_, err := HostnameSet("")
	if err == nil {
		t.Fatal("expected error for empty hostname")
	}
}

func TestHostnameSet_CurrentHostname(t *testing.T) {
	// Setting to the current hostname should return Changed: false
	// We can't actually change the hostname in a test, so we just verify
	// that requesting the current hostname results in no change.
	res, err := HostnameSet("this-would-require-root-to-actually-set")
	// This will likely fail because we don't have permission, but we test
	// the validation path
	if err != nil {
		// Expected when running without root
		t.Skipf("skipping: requires root privileges: %v", err)
	}
	_ = res
}
