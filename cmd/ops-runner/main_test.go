package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/security"
	"github.com/j4ckzh0u/opslang/internal/runner"
)

func testKeyPairE2E(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	return pub, priv
}

func writeRawPubKey(t *testing.T, pub ed25519.PublicKey) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "runner.pub")
	if err := os.WriteFile(path, pub, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func signedHostnamePackage(t *testing.T, priv ed25519.PrivateKey) []byte {
	t.Helper()
	pkg := &runner.InstructionPackage{
		Version:   "1.0",
		TaskID:    "e2e-sign",
		Privilege: string(ast.PrivilegeReadOnly),
		Instructions: []runner.Instruction{
			{Op: "sys.hostname", Args: map[string]interface{}{}, Assign: "h"},
		},
	}
	if err := runner.SignPackage(pkg, priv); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// tamperJSON flips a field in an already-signed package's JSON, exactly
// what an attacker in the delivery path would do.
func tamperJSON(t *testing.T, data []byte) []byte {
	t.Helper()
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["dry_run"] = true
	out, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// resetRunnerFlags isolates the package-level flag vars between tests.
func resetRunnerFlags(t *testing.T) {
	t.Helper()
	oldPath, oldStatus := pubKeyPath, status
	t.Cleanup(func() { pubKeyPath = oldPath; status = oldStatus })
	pubKeyPath, status = "", ""
}

func TestRunRejectsUnsignedWhenEnforced(t *testing.T) {
	resetRunnerFlags(t)
	_, priv := testKeyPairE2E(t)
	pubKeyPath = writeRawPubKey(t, priv.Public().(ed25519.PublicKey))

	var out bytes.Buffer
	err := run(bytes.NewReader([]byte(`{"version":"1.0","instructions":[]}`)), &out)
	if err != nil {
		t.Fatalf("unsigned package must be a structured rejection, not usage error: %v", err)
	}
	if status != "failed" {
		t.Fatalf("status = %q, want failed", status)
	}
	var output runner.Output
	if uerr := json.Unmarshal(out.Bytes(), &output); uerr != nil {
		t.Fatalf("rejection output must be valid JSON: %v (%q)", uerr, out.String())
	}
	if len(output.Errors) == 0 || !strings.Contains(output.Errors[0], "signature") {
		t.Fatalf("rejection should mention signature, got: %v", output.Errors)
	}
}

func TestRunAcceptsValidlySignedPackage(t *testing.T) {
	resetRunnerFlags(t)
	_, priv := testKeyPairE2E(t)
	pubKeyPath = writeRawPubKey(t, priv.Public().(ed25519.PublicKey))

	var out bytes.Buffer
	if err := run(bytes.NewReader(signedHostnamePackage(t, priv)), &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if status != "ok" {
		t.Fatalf("status = %q, want ok (output: %s)", status, out.String())
	}
}

func TestRunDetectsTamperedPackage(t *testing.T) {
	resetRunnerFlags(t)
	_, priv := testKeyPairE2E(t)
	pubKeyPath = writeRawPubKey(t, priv.Public().(ed25519.PublicKey))

	payload := tamperJSON(t, signedHostnamePackage(t, priv))
	var out bytes.Buffer
	if err := run(bytes.NewReader(payload), &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if status != "failed" {
		t.Fatalf("tampered package executed with status %q; enforcement failed", status)
	}
	if !strings.Contains(out.String(), "signature") {
		t.Fatalf("failure should mention signature: %s", out.String())
	}
}

func TestRunMissingKeyFileIsUsageError(t *testing.T) {
	resetRunnerFlags(t)
	pubKeyPath = filepath.Join(t.TempDir(), "does-not-exist.pub")

	var out bytes.Buffer
	err := run(bytes.NewReader([]byte(`{"version":"1.0","instructions":[]}`)), &out)
	if err == nil {
		t.Fatal("missing public key file must be an error (exit 3 path)")
	}
	if !strings.Contains(err.Error(), "public key") {
		t.Fatalf("error should mention the public key file: %v", err)
	}
}

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
