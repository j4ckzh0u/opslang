//go:build opssec

package main

// The hex-encoded public key acceptance path only exists in opssec builds,
// so its test lives under the same tag as the code it exercises.

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/security"
)

func TestRunHexEncodedPubKeyAccepted(t *testing.T) {
	resetRunnerFlags(t)
	_, priv := testKeyPairE2E(t)
	hexPath := filepath.Join(t.TempDir(), "runner-hex.pub")
	hexPub := security.PublicKeyToString(priv.Public().(ed25519.PublicKey))
	if err := os.WriteFile(hexPath, []byte(hexPub+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	pubKeyPath = hexPath

	var out bytes.Buffer
	if err := run(bytes.NewReader(signedHostnamePackage(t, priv)), &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if status != "ok" {
		t.Fatalf("status = %q, want ok with hex-encoded key (output: %s)", status, out.String())
	}
}
