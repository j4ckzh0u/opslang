package openssl_csr

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateValidation(t *testing.T) {
	// Missing common_name
	r := Generate(CSRConfig{
		KeyFile:    "test.key",
		OutputFile: "test.csr",
	})
	if r.Success {
		t.Error("expected failure for missing common_name")
	}

	// Missing key_file
	r = Generate(CSRConfig{
		CommonName: "test.example.com",
		OutputFile: "test.csr",
	})
	if r.Success {
		t.Error("expected failure for missing key_file")
	}

	// Missing output_file
	r = Generate(CSRConfig{
		CommonName: "test.example.com",
		KeyFile:    "test.key",
	})
	if r.Success {
		t.Error("expected failure for missing output_file")
	}
}

func TestGenerateSuccess(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "test.key")
	csrFile := filepath.Join(tmpDir, "test.csr")

	// Generate a test key first
	err := generateTestKey(t, keyFile)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	// Generate CSR
	r := Generate(CSRConfig{
		CommonName:   "test.example.com",
		Organization: "Test Org",
		Country:      "US",
		DNSNames:     []string{"test.example.com", "www.example.com"},
		KeyFile:      keyFile,
		OutputFile:   csrFile,
	})

	if !r.Success {
		t.Errorf("expected success: %v", r.Error)
	}
	if !r.Changed {
		t.Error("expected changed to be true")
	}

	// Verify file exists
	if _, err := os.Stat(csrFile); os.IsNotExist(err) {
		t.Error("CSR file was not created")
	}

	// Test idempotency (without force)
	r2 := Generate(CSRConfig{
		CommonName: "test.example.com",
		KeyFile:    keyFile,
		OutputFile: csrFile,
	})
	if !r2.Success {
		t.Error("expected success for idempotent call")
	}
	if r2.Changed {
		t.Error("expected changed to be false for idempotent call")
	}
}

func TestInfoSuccess(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "test.key")
	csrFile := filepath.Join(tmpDir, "test.csr")

	// Generate test key and CSR
	generateTestKey(t, keyFile)
	Generate(CSRConfig{
		CommonName: "test.example.com",
		KeyFile:    keyFile,
		OutputFile: csrFile,
	})

	// Get info
	r := Info(csrFile)
	if !r.Success {
		t.Errorf("expected success: %v", r.Error)
	}
}

func TestInfoNonExistent(t *testing.T) {
	r := Info("/nonexistent/file.csr")
	if r.Success {
		t.Error("expected failure for non-existent file")
	}
}

func TestDeleteSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	csrFile := filepath.Join(tmpDir, "test.csr")

	// Create a dummy file
	os.WriteFile(csrFile, []byte("test"), 0644)

	r := Delete(csrFile)
	if !r.Success {
		t.Errorf("expected success: %v", r.Error)
	}
	if !r.Changed {
		t.Error("expected changed to be true")
	}

	// Verify file is deleted
	if _, err := os.Stat(csrFile); !os.IsNotExist(err) {
		t.Error("file was not deleted")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	r := Delete("/nonexistent/file.csr")
	if !r.Success {
		t.Error("expected success even for non-existent file")
	}
	if r.Changed {
		t.Error("expected changed to be false for non-existent file")
	}
}

// Helper function to generate a test key
func generateTestKey(t *testing.T, keyFile string) error {
	t.Helper()
	// Generate a real ECDSA key for testing
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
