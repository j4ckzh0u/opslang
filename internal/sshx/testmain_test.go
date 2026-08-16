package sshx

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain isolates TOFU known-hosts state for every test in this package:
// mock servers present a fresh key per run and must neither read nor
// pollute the developer's real known-hosts file.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "opslang-test-kh-")
	if err != nil {
		os.Exit(1)
	}
	defer os.RemoveAll(dir)
	os.Setenv("OPSLANG_KNOWN_HOSTS", filepath.Join(dir, "known_hosts"))
	os.Exit(m.Run())
}
