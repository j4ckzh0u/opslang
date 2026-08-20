package openssl_privatekey

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRSA(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_rsa.key")

	result := Generate(GenerateConfig{
		Path: path,
		Type: "rsa",
		Size: 2048,
	})

	if !result.Success {
		t.Fatalf("Generate() failed: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true")
	}
	if result.Type != "rsa" {
		t.Errorf("expected type=rsa, got %s", result.Type)
	}
	if result.Size != 2048 {
		t.Errorf("expected size=2048, got %d", result.Size)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("key file not created: %v", err)
	}

	// Verify file permissions
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestGenerateECDSA(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_ecdsa.key")

	result := Generate(GenerateConfig{
		Path: path,
		Type: "ecdsa",
		Size: 256,
	})

	if !result.Success {
		t.Fatalf("Generate() failed: %v", result.Error)
	}
	if result.Type != "ecdsa" {
		t.Errorf("expected type=ecdsa, got %s", result.Type)
	}
}

func TestGenerateEd25519(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_ed25519.key")

	result := Generate(GenerateConfig{
		Path: path,
		Type: "ed25519",
	})

	if !result.Success {
		t.Fatalf("Generate() failed: %v", result.Error)
	}
	if result.Type != "ed25519" {
		t.Errorf("expected type=ed25519, got %s", result.Type)
	}
}

func TestGenerateDefaultType(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_default.key")

	result := Generate(GenerateConfig{
		Path: path,
	})

	if !result.Success {
		t.Fatalf("Generate() failed: %v", result.Error)
	}
	if result.Type != "rsa" {
		t.Errorf("expected default type=rsa, got %s", result.Type)
	}
}

func TestGenerateNoChange(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_nochange.key")

	// Create key first
	Generate(GenerateConfig{Path: path})

	// Try to generate again without force
	result := Generate(GenerateConfig{
		Path: path,
	})

	if !result.Success {
		t.Fatalf("Generate() failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for existing key")
	}
}

func TestGenerateForce(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_force.key")

	// Create key first
	Generate(GenerateConfig{Path: path})

	// Force regenerate
	result := Generate(GenerateConfig{
		Path:  path,
		Force: true,
	})

	if !result.Success {
		t.Fatalf("Generate() failed: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true with force=true")
	}
}

func TestGenerateInvalidType(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_invalid.key")

	result := Generate(GenerateConfig{
		Path: path,
		Type: "invalid",
	})

	if result.Success {
		t.Error("expected failure for invalid key type")
	}
	if !strings.Contains(result.Error, "unsupported key type") {
		t.Errorf("expected 'unsupported key type' error, got: %v", result.Error)
	}
}

func TestGenerateEmptyPath(t *testing.T) {
	result := Generate(GenerateConfig{})

	if result.Success {
		t.Error("expected failure for empty path")
	}
}

func TestInfo(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_info.key")

	// Create key
	Generate(GenerateConfig{
		Path: path,
		Type: "rsa",
		Size: 2048,
	})

	// Get info
	info := Info(path)

	if !info.Success {
		t.Fatalf("Info() failed: %v", info.Error)
	}
	if info.Type != "rsa" {
		t.Errorf("expected type=rsa, got %s", info.Type)
	}
	if info.Size != 2048 {
		t.Errorf("expected size=2048, got %d", info.Size)
	}
}

func TestInfoECDSA(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_info_ecdsa.key")

	Generate(GenerateConfig{
		Path: path,
		Type: "ecdsa",
		Size: 384,
	})

	info := Info(path)

	if !info.Success {
		t.Fatalf("Info() failed: %v", info.Error)
	}
	if info.Type != "ecdsa" {
		t.Errorf("expected type=ecdsa, got %s", info.Type)
	}
	if info.Size != 384 {
		t.Errorf("expected size=384, got %d", info.Size)
	}
}

func TestInfoNonExistent(t *testing.T) {
	info := Info("/nonexistent/path.key")

	if info.Success {
		t.Error("expected failure for non-existent file")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_delete.key")

	// Create key
	Generate(GenerateConfig{Path: path})

	// Delete
	result := Delete(path)

	if !result.Success {
		t.Fatalf("Delete() failed: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true")
	}

	// Verify file is gone
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	result := Delete("/nonexistent/path.key")

	if result.Success {
		t.Error("expected failure for non-existent file")
	}
}
