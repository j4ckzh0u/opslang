//go:build opssec

package security

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
)

func TestNewSignatureManagerGenerateKey(t *testing.T) {
	sm := NewSignatureManager(nil)
	if sm.privateKey == nil {
		t.Fatal("privateKey is nil")
	}
	if sm.publicKey == nil {
		t.Fatal("publicKey is nil")
	}
	if len(sm.privateKey) != ed25519.PrivateKeySize {
		t.Errorf("privateKey size = %d, want %d", len(sm.privateKey), ed25519.PrivateKeySize)
	}
	if len(sm.publicKey) != ed25519.PublicKeySize {
		t.Errorf("publicKey size = %d, want %d", len(sm.publicKey), ed25519.PublicKeySize)
	}
}

func TestNewSignatureManagerWithKey(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	sm := NewSignatureManager(priv)

	if !sm.publicKey.Equal(pub) {
		t.Error("public key mismatch")
	}
	if !sm.privateKey.Equal(priv) {
		t.Error("private key mismatch")
	}
}

func TestSignAndVerify(t *testing.T) {
	sm := NewSignatureManager(nil)
	data := []byte("hello world")

	sig := sm.Sign(data)
	if len(sig) != ed25519.SignatureSize {
		t.Errorf("signature size = %d, want %d", len(sig), ed25519.SignatureSize)
	}

	if !sm.Verify(data, sig) {
		t.Error("Verify should return true for valid signature")
	}

	// Tampered data
	if sm.Verify([]byte("tampered"), sig) {
		t.Error("Verify should return false for tampered data")
	}

	// Tampered signature
	badSig := make([]byte, len(sig))
	copy(badSig, sig)
	badSig[0] ^= 0xff
	if sm.Verify(data, badSig) {
		t.Error("Verify should return false for tampered signature")
	}
}

func TestPublicKeyAndPrivateKeyAccessors(t *testing.T) {
	sm := NewSignatureManager(nil)
	if sm.PublicKey() == nil {
		t.Error("PublicKey() returned nil")
	}
	if sm.PrivateKey() == nil {
		t.Error("PrivateKey() returned nil")
	}
}

func TestSignFileAndVerifyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.bin")
	content := []byte("file content to sign")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	sm := NewSignatureManager(nil)

	sig, err := sm.SignFile(filePath)
	if err != nil {
		t.Fatalf("SignFile error: %v", err)
	}

	valid, err := sm.VerifyFile(filePath, sig)
	if err != nil {
		t.Fatalf("VerifyFile error: %v", err)
	}
	if !valid {
		t.Error("VerifyFile should return true")
	}
}

func TestSignFileNotFound(t *testing.T) {
	sm := NewSignatureManager(nil)
	_, err := sm.SignFile("/nonexistent/file")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestVerifyFileNotFound(t *testing.T) {
	sm := NewSignatureManager(nil)
	_, err := sm.VerifyFile("/nonexistent/file", []byte("sig"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestSaveAndLoadPublicKey(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSignatureManager(nil)

	path := filepath.Join(tmpDir, "pub.key")
	if err := sm.SavePublicKey(path); err != nil {
		t.Fatalf("SavePublicKey error: %v", err)
	}

	loaded, err := LoadPublicKey(path)
	if err != nil {
		t.Fatalf("LoadPublicKey error: %v", err)
	}
	if !loaded.Equal(sm.PublicKey()) {
		t.Error("loaded public key mismatch")
	}
}

func TestLoadPublicKeyInvalidSize(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.key")
	if err := os.WriteFile(path, []byte("tooshort"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadPublicKey(path)
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestLoadPublicKeyNotFound(t *testing.T) {
	_, err := LoadPublicKey("/nonexistent/key")
	if err == nil {
		t.Error("expected error")
	}
}

func TestSaveAndLoadPrivateKey(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSignatureManager(nil)

	path := filepath.Join(tmpDir, "priv.key")
	if err := sm.SavePrivateKey(path); err != nil {
		t.Fatalf("SavePrivateKey error: %v", err)
	}

	loaded, err := LoadPrivateKey(path)
	if err != nil {
		t.Fatalf("LoadPrivateKey error: %v", err)
	}
	if !loaded.Equal(sm.PrivateKey()) {
		t.Error("loaded private key mismatch")
	}
}

func TestLoadPrivateKeyInvalidSize(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.key")
	if err := os.WriteFile(path, []byte("tooshort"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadPrivateKey(path)
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestLoadPrivateKeyNotFound(t *testing.T) {
	_, err := LoadPrivateKey("/nonexistent/key")
	if err == nil {
		t.Error("expected error")
	}
}

func TestSignatureToString(t *testing.T) {
	sm := NewSignatureManager(nil)
	sig := sm.Sign([]byte("data"))

	hexStr := SignatureToString(sig)
	if hexStr == "" {
		t.Error("SignatureToString returned empty string")
	}

	recovered, err := StringToSignature(hexStr)
	if err != nil {
		t.Fatalf("StringToSignature error: %v", err)
	}
	if len(recovered) != len(sig) {
		t.Errorf("recovered length = %d, want %d", len(recovered), len(sig))
	}
	for i := range sig {
		if sig[i] != recovered[i] {
			t.Errorf("byte %d mismatch", i)
			break
		}
	}
}

func TestStringToSignatureInvalidHex(t *testing.T) {
	_, err := StringToSignature("not-hex!")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestPublicKeyToString(t *testing.T) {
	sm := NewSignatureManager(nil)
	hexStr := PublicKeyToString(sm.PublicKey())
	if hexStr == "" {
		t.Error("PublicKeyToString returned empty string")
	}

	recovered, err := StringToPublicKey(hexStr)
	if err != nil {
		t.Fatalf("StringToPublicKey error: %v", err)
	}
	if !recovered.Equal(sm.PublicKey()) {
		t.Error("recovered public key mismatch")
	}
}

func TestStringToPublicKeyInvalidHex(t *testing.T) {
	_, err := StringToPublicKey("not-hex!")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestStringToPublicKeyInvalidSize(t *testing.T) {
	// Valid hex but wrong size (2 bytes)
	_, err := StringToPublicKey("aabb")
	if err == nil {
		t.Error("expected error for wrong size")
	}
}

func TestComputeChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "data.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	sum1, err := ComputeChecksum(path)
	if err != nil {
		t.Fatalf("ComputeChecksum error: %v", err)
	}
	if sum1 == "" {
		t.Error("checksum is empty")
	}
	if len(sum1) != 64 { // SHA256 hex = 64 chars
		t.Errorf("checksum length = %d, want 64", len(sum1))
	}

	// Same content should produce same checksum
	sum2, err := ComputeChecksum(path)
	if err != nil {
		t.Fatal(err)
	}
	if sum1 != sum2 {
		t.Errorf("checksums differ: %q vs %q", sum1, sum2)
	}

	// Different content should produce different checksum
	if err := os.WriteFile(path, []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}
	sum3, err := ComputeChecksum(path)
	if err != nil {
		t.Fatal(err)
	}
	if sum1 == sum3 {
		t.Error("checksums should differ for different content")
	}
}

func TestComputeChecksumFileNotFound(t *testing.T) {
	_, err := ComputeChecksum("/nonexistent/file")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCrossKeyVerification(t *testing.T) {
	// Sign with one key, verify with another should fail
	sm1 := NewSignatureManager(nil)
	sm2 := NewSignatureManager(nil)

	data := []byte("cross-key test")
	sig := sm1.Sign(data)

	if sm2.Verify(data, sig) {
		t.Error("verification with different key should fail")
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSignatureManager(nil)

	pubPath := filepath.Join(tmpDir, "pub.key")
	privPath := filepath.Join(tmpDir, "priv.key")

	if err := sm.SavePublicKey(pubPath); err != nil {
		t.Fatal(err)
	}
	if err := sm.SavePrivateKey(privPath); err != nil {
		t.Fatal(err)
	}

	loadedPriv, err := LoadPrivateKey(privPath)
	if err != nil {
		t.Fatal(err)
	}

	// Create new manager from loaded key and verify it can verify signatures
	sm2 := NewSignatureManager(loadedPriv)
	data := []byte("roundtrip test")
	sig := sm.Sign(data)
	if !sm2.Verify(data, sig) {
		t.Error("loaded key should verify original signature")
	}

	sig2 := sm2.Sign(data)
	if !sm.Verify(data, sig2) {
		t.Error("original key should verify loaded key's signature")
	}
}
