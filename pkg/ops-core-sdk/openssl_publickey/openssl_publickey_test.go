package openssl_publickey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractValidation(t *testing.T) {
	// Missing private_key_file
	r := Extract("", "output.pub", false)
	if r.Success {
		t.Error("expected failure for missing private_key_file")
	}

	// Missing output_file
	r = Extract("input.key", "", false)
	if r.Success {
		t.Error("expected failure for missing output_file")
	}
}

func TestExtractRSA(t *testing.T) {
	tmpDir := t.TempDir()
	privKeyFile := filepath.Join(tmpDir, "test.key")
	pubKeyFile := filepath.Join(tmpDir, "test.pub")

	// Generate RSA private key
	err := generateRSAPrivateKey(t, privKeyFile)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Extract public key
	r := Extract(privKeyFile, pubKeyFile, false)
	if !r.Success {
		t.Errorf("expected success: %v", r.Error)
	}
	if !r.Changed {
		t.Error("expected changed to be true")
	}

	// Verify file exists
	if _, err := os.Stat(pubKeyFile); os.IsNotExist(err) {
		t.Error("public key file was not created")
	}

	// Verify it's a valid public key
	keyType, err := GetKeyType(pubKeyFile)
	if err != nil {
		t.Errorf("failed to get key type: %v", err)
	}
	if keyType != "RSA" {
		t.Errorf("expected RSA, got %s", keyType)
	}

	// Test idempotency
	r2 := Extract(privKeyFile, pubKeyFile, false)
	if !r2.Success {
		t.Error("expected success for idempotent call")
	}
	if r2.Changed {
		t.Error("expected changed to be false for idempotent call")
	}
}

func TestExtractECDSA(t *testing.T) {
	tmpDir := t.TempDir()
	privKeyFile := filepath.Join(tmpDir, "test.key")
	pubKeyFile := filepath.Join(tmpDir, "test.pub")

	// Generate ECDSA private key
	err := generateECDSAPrivateKey(t, privKeyFile)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Extract public key
	r := Extract(privKeyFile, pubKeyFile, false)
	if !r.Success {
		t.Errorf("expected success: %v", r.Error)
	}

	// Verify it's an ECDSA key
	keyType, err := GetKeyType(pubKeyFile)
	if err != nil {
		t.Errorf("failed to get key type: %v", err)
	}
	if keyType != "ECDSA-P256" {
		t.Errorf("expected ECDSA-P256, got %s", keyType)
	}
}

func TestInfoSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	privKeyFile := filepath.Join(tmpDir, "test.key")
	pubKeyFile := filepath.Join(tmpDir, "test.pub")

	generateRSAPrivateKey(t, privKeyFile)
	Extract(privKeyFile, pubKeyFile, false)

	r := Info(pubKeyFile)
	if !r.Success {
		t.Errorf("expected success: %v", r.Error)
	}
}

func TestInfoNonExistent(t *testing.T) {
	r := Info("/nonexistent/file.pub")
	if r.Success {
		t.Error("expected failure for non-existent file")
	}
}

func TestDeleteSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	pubKeyFile := filepath.Join(tmpDir, "test.pub")

	os.WriteFile(pubKeyFile, []byte("test"), 0644)

	r := Delete(pubKeyFile)
	if !r.Success {
		t.Errorf("expected success: %v", r.Error)
	}
	if !r.Changed {
		t.Error("expected changed to be true")
	}

	if _, err := os.Stat(pubKeyFile); !os.IsNotExist(err) {
		t.Error("file was not deleted")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	r := Delete("/nonexistent/file.pub")
	if !r.Success {
		t.Error("expected success even for non-existent file")
	}
	if r.Changed {
		t.Error("expected changed to be false for non-existent file")
	}
}

// Helper functions
func generateRSAPrivateKey(t *testing.T, keyFile string) error {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	return os.WriteFile(keyFile, keyPEM, 0600)
}

func generateECDSAPrivateKey(t *testing.T, keyFile string) error {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	return os.WriteFile(keyFile, keyPEM, 0600)
}
