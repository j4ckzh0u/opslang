package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	pub, priv, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if len(pub) == 0 {
		t.Error("Public key is empty")
	}
	if len(priv) == 0 {
		t.Error("Private key is empty")
	}
}

func TestSignAndVerify(t *testing.T) {
	pub, priv, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	payload := []byte("hello world")

	// Sign
	sig := Sign(payload, priv)
	if len(sig) == 0 {
		t.Fatal("Signature is empty")
	}

	// Verify with correct key
	if !Verify(payload, sig, pub) {
		t.Error("Verify failed with correct key")
	}

	// Verify with wrong payload
	if Verify([]byte("wrong payload"), sig, pub) {
		t.Error("Verify succeeded with wrong payload")
	}

	// Verify with wrong key
	pub2, _, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if Verify(payload, sig, pub2) {
		t.Error("Verify succeeded with wrong key")
	}
}

func TestSignAndVerifyEmptyPayload(t *testing.T) {
	pub, priv, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	payload := []byte{}
	sig := Sign(payload, priv)

	if !Verify(payload, sig, pub) {
		t.Error("Verify failed for empty payload")
	}
}

func TestSaveAndLoadPublicKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "pub.pem")

	pub, _, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Save
	if err := SavePublicKey(keyPath, pub); err != nil {
		t.Fatalf("SavePublicKey failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatal("Public key file not created")
	}

	// Load
	loaded, err := LoadPublicKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPublicKey failed: %v", err)
	}

	// Compare
	if len(loaded) != len(pub) {
		t.Errorf("Loaded key length = %d, want %d", len(loaded), len(pub))
	}
	for i := range loaded {
		if loaded[i] != pub[i] {
			t.Error("Loaded key does not match original")
			break
		}
	}
}

func TestSaveAndLoadPrivateKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "priv.pem")

	_, priv, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Save
	if err := SavePrivateKey(keyPath, priv); err != nil {
		t.Fatalf("SavePrivateKey failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatal("Private key file not created")
	}

	// Verify file permissions (should be 0600)
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("File permissions = %o, want 0600", perm)
	}

	// Load
	loaded, err := LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}

	// Compare
	if len(loaded) != len(priv) {
		t.Errorf("Loaded key length = %d, want %d", len(loaded), len(priv))
	}
	for i := range loaded {
		if loaded[i] != priv[i] {
			t.Error("Loaded key does not match original")
			break
		}
	}
}

func TestSavePublicKeyCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "nested", "dir", "pub.pem")

	pub, _, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if err := SavePublicKey(keyPath, pub); err != nil {
		t.Fatalf("SavePublicKey failed: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(keyPath)); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestLoadPublicKeyInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid.pem")

	// Write invalid content
	if err := os.WriteFile(keyPath, []byte("not a pem file"), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	_, err := LoadPublicKey(keyPath)
	if err == nil {
		t.Error("Expected error loading invalid public key file")
	}
}

func TestLoadPrivateKeyInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid.pem")

	// Write invalid content
	if err := os.WriteFile(keyPath, []byte("not a pem file"), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	_, err := LoadPrivateKey(keyPath)
	if err == nil {
		t.Error("Expected error loading invalid private key file")
	}
}

func TestLoadPublicKeyNonExistent(t *testing.T) {
	_, err := LoadPublicKey("/nonexistent/path/pub.pem")
	if err == nil {
		t.Error("Expected error loading non-existent file")
	}
}

func TestLoadPrivateKeyNonExistent(t *testing.T) {
	_, err := LoadPrivateKey("/nonexistent/path/priv.pem")
	if err == nil {
		t.Error("Expected error loading non-existent file")
	}
}

func TestFullKeyRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	pubPath := filepath.Join(tmpDir, "pub.pem")
	privPath := filepath.Join(tmpDir, "priv.pem")

	// Generate keys
	pub, priv, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Save keys
	if err := SavePublicKey(pubPath, pub); err != nil {
		t.Fatalf("SavePublicKey failed: %v", err)
	}
	if err := SavePrivateKey(privPath, priv); err != nil {
		t.Fatalf("SavePrivateKey failed: %v", err)
	}

	// Load keys
	loadedPub, err := LoadPublicKey(pubPath)
	if err != nil {
		t.Fatalf("LoadPublicKey failed: %v", err)
	}
	loadedPriv, err := LoadPrivateKey(privPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}

	// Sign with loaded private key
	payload := []byte("test message for round-trip")
	sig := Sign(payload, loadedPriv)

	// Verify with loaded public key
	if !Verify(payload, sig, loadedPub) {
		t.Error("Verify failed with loaded keys")
	}
}

func TestLoadPublicKeyWrongType(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "wrong_type.pem")

	// Write a PEM with wrong type
	pemContent := []byte(`-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHBfpEgcMFvMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBnRl
c3RjYTAeFw0yNDAxMDEwMDAwMDBaFw0yNTAxMDEwMDAwMDBaMBExDzANBgNVBAMM
BnRlc3RjYTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC7o96FCk7t4kLtGuAIx+Mn
d9g3X7Z6Ji+Q4W1Y9V6n+R3y9C7V6G9p1V2G2b7Q7L5p9T5v8O9Y9V6n+R3y9C7V
AgMBAAEwDQYJKoZIhvcNAQELBQADQQBp9V6n+R3y9C7V6G9p1V2G2b7Q7L5p9T5v
-----END CERTIFICATE-----`)
	if err := os.WriteFile(keyPath, pemContent, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err := LoadPublicKey(keyPath)
	if err == nil {
		t.Error("Expected error loading wrong type PEM")
	}
}
