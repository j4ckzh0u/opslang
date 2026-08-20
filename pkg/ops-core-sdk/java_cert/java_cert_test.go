package java_cert

import (
	"testing"
)

// TestImportValidation tests input validation.
func TestImportValidation(t *testing.T) {
	_, err := Import("", "pass", "alias", "/tmp/cert.pem", "")
	if err == nil {
		t.Fatal("expected error for empty keystore_path, got nil")
	}
	_, err = Import("/tmp/ks.jks", "pass", "alias", "", "")
	if err == nil {
		t.Fatal("expected error for empty cert_path, got nil")
	}
}

// TestRemoveValidation tests input validation.
func TestRemoveValidation(t *testing.T) {
	_, err := Remove("", "pass", "alias")
	if err == nil {
		t.Fatal("expected error for empty keystore_path, got nil")
	}
	_, err = Remove("/tmp/ks.jks", "pass", "")
	if err == nil {
		t.Fatal("expected error for empty alias, got nil")
	}
}

// TestListValidation tests input validation.
func TestListValidation(t *testing.T) {
	_, err := List("", "pass")
	if err == nil {
		t.Fatal("expected error for empty keystore_path, got nil")
	}
}

// TestExistsValidation tests input validation.
func TestExistsValidation(t *testing.T) {
	_, err := Exists("", "pass", "alias")
	if err == nil {
		t.Fatal("expected error for empty keystore_path, got nil")
	}
	_, err = Exists("/tmp/ks.jks", "pass", "")
	if err == nil {
		t.Fatal("expected error for empty alias, got nil")
	}
}

// TestExportValidation tests input validation.
func TestExportValidation(t *testing.T) {
	_, err := Export("", "pass", "alias", "/tmp/out.pem", "")
	if err == nil {
		t.Fatal("expected error for empty keystore_path, got nil")
	}
	_, err = Export("/tmp/ks.jks", "pass", "", "/tmp/out.pem", "")
	if err == nil {
		t.Fatal("expected error for empty alias, got nil")
	}
	_, err = Export("/tmp/ks.jks", "pass", "alias", "", "")
	if err == nil {
		t.Fatal("expected error for empty output_path, got nil")
	}
}

// TestInfoValidation tests input validation.
func TestInfoValidation(t *testing.T) {
	_, err := Info("", "pass")
	if err == nil {
		t.Fatal("expected error for empty keystore_path, got nil")
	}
}

// TestImportChainValidation tests input validation.
func TestImportChainValidation(t *testing.T) {
	_, err := ImportChain("", "pass", "/tmp/cert.p12", "pass")
	if err == nil {
		t.Fatal("expected error for empty keystore_path, got nil")
	}
	_, err = ImportChain("/tmp/ks.jks", "pass", "", "pass")
	if err == nil {
		t.Fatal("expected error for empty p12_path, got nil")
	}
}

// TestChangePasswordValidation tests input validation.
func TestChangePasswordValidation(t *testing.T) {
	_, err := ChangePassword("", "old", "new")
	if err == nil {
		t.Fatal("expected error for empty keystore_path, got nil")
	}
	_, err = ChangePassword("/tmp/ks.jks", "old", "")
	if err == nil {
		t.Fatal("expected error for empty new_password, got nil")
	}
}

// TestParseKeystoreList tests the keystore list parser.
func TestParseKeystoreList(t *testing.T) {
	output := `Alias name: mycert
Entry type: trustedCertEntry
SHA256: AB:CD:EF:12:34

Alias name: othercert
Entry type: privateKeyEntry
SHA1: 11:22:33:44`

	certs := parseKeystoreList(output)
	if len(certs) != 2 {
		t.Fatalf("expected 2 certs, got %d", len(certs))
	}
	if certs[0].Alias != "mycert" {
		t.Errorf("expected alias 'mycert', got %q", certs[0].Alias)
	}
	if certs[0].Type != "trustedCertEntry" {
		t.Errorf("expected type 'trustedCertEntry', got %q", certs[0].Type)
	}
	if certs[1].Alias != "othercert" {
		t.Errorf("expected alias 'othercert', got %q", certs[1].Alias)
	}
}
