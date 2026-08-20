// Package java_cert provides Java keystore management operations.
// Equivalent to Ansible's java_cert module.
package java_cert

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result represents a generic keystore operation result.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CertInfo represents information about a certificate in a keystore.
type CertInfo struct {
	Alias       string `json:"alias"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Valid       bool   `json:"valid"`
}

// KeystoreInfo represents keystore metadata.
type KeystoreInfo struct {
	Status  string     `json:"status"`
	Path    string     `json:"path"`
	Type    string     `json:"type"`
	Certs   []CertInfo `json:"certs,omitempty"`
	Error   string     `json:"error,omitempty"`
}

func findKeytool() (string, error) {
	if p, err := exec.LookPath("keytool"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("keytool not found in PATH")
}

func runKeytool(args ...string) (string, error) {
	kt, err := findKeytool()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(kt, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Import imports a certificate into a Java keystore.
func Import(keystorePath string, password string, alias string, certPath string, certType string) (Result, error) {
	if keystorePath == "" || certPath == "" {
		return Result{Status: "failed", Error: "keystore_path and cert_path are required"}, fmt.Errorf("keystore_path and cert_path are required")
	}
	if alias == "" {
		alias = "imported"
	}
	if certType == "" {
		certType = "x509"
	}
	args := []string{
		"-importcert",
		"-keystore", keystorePath,
		"-storepass", password,
		"-alias", alias,
		"-file", certPath,
		"-noprompt",
	}
	if certType != "x509" {
		args = append(args, "-certtype", certType)
	}
	out, err := runKeytool(args...)
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("import cert: %v: %s", err, out)}, err
	}
	return Result{Status: "success", Changed: true, Message: out}, nil
}

// Remove removes a certificate from a Java keystore.
func Remove(keystorePath string, password string, alias string) (Result, error) {
	if keystorePath == "" || alias == "" {
		return Result{Status: "failed", Error: "keystore_path and alias are required"}, fmt.Errorf("keystore_path and alias are required")
	}
	args := []string{
		"-delete",
		"-keystore", keystorePath,
		"-storepass", password,
		"-alias", alias,
	}
	out, err := runKeytool(args...)
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("remove cert: %v: %s", err, out)}, err
	}
	return Result{Status: "success", Changed: true, Message: out}, nil
}

// List lists all certificates in a keystore.
func List(keystorePath string, password string) ([]CertInfo, error) {
	if keystorePath == "" {
		return nil, fmt.Errorf("keystore_path is required")
	}
	args := []string{
		"-list",
		"-keystore", keystorePath,
		"-storepass", password,
		"-v",
	}
	out, err := runKeytool(args...)
	if err != nil {
		return nil, fmt.Errorf("list certs: %w: %s", err, out)
	}
	return parseKeystoreList(out), nil
}

// Exists checks if a certificate alias exists in the keystore.
func Exists(keystorePath string, password string, alias string) (bool, error) {
	if keystorePath == "" || alias == "" {
		return false, fmt.Errorf("keystore_path and alias are required")
	}
	args := []string{
		"-list",
		"-keystore", keystorePath,
		"-storepass", password,
		"-alias", alias,
	}
	_, err := runKeytool(args...)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Export exports a certificate from a keystore to a file.
func Export(keystorePath string, password string, alias string, outputPath string, certType string) (Result, error) {
	if keystorePath == "" || alias == "" || outputPath == "" {
		return Result{Status: "failed", Error: "keystore_path, alias, and output_path are required"}, fmt.Errorf("keystore_path, alias, and output_path are required")
	}
	if certType == "" {
		certType = "x509"
	}
	args := []string{
		"-exportcert",
		"-keystore", keystorePath,
		"-storepass", password,
		"-alias", alias,
		"-file", outputPath,
		"-rfc",
	}
	out, err := runKeytool(args...)
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("export cert: %v: %s", err, out)}, err
	}
	return Result{Status: "success", Changed: true, Message: out}, nil
}

// Info returns metadata about a keystore.
func Info(keystorePath string, password string) (KeystoreInfo, error) {
	if keystorePath == "" {
		return KeystoreInfo{Status: "failed", Error: "keystore_path is required"}, fmt.Errorf("keystore_path is required")
	}
	info := KeystoreInfo{Status: "success", Path: keystorePath, Type: "jks"}
	args := []string{
		"-list",
		"-keystore", keystorePath,
		"-storepass", password,
		"-v",
	}
	out, err := runKeytool(args...)
	if err != nil {
		info.Status = "failed"
		info.Error = fmt.Sprintf("keystore info: %v: %s", err, out)
		return info, err
	}
	info.Certs = parseKeystoreList(out)
	return info, nil
}

// ImportChain imports a PKCS12 certificate chain into a keystore.
func ImportChain(keystorePath string, password string, p12Path string, p12Password string) (Result, error) {
	if keystorePath == "" || p12Path == "" {
		return Result{Status: "failed", Error: "keystore_path and p12_path are required"}, fmt.Errorf("keystore_path and p12_path are required")
	}
	args := []string{
		"-importkeystore",
		"-destkeystore", keystorePath,
		"-deststorepass", password,
		"-srckeystore", p12Path,
		"-srcstorepass", p12Password,
		"-srcstoretype", "PKCS12",
		"-noprompt",
	}
	out, err := runKeytool(args...)
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("import chain: %v: %s", err, out)}, err
	}
	return Result{Status: "success", Changed: true, Message: out}, nil
}

// ChangePassword changes the password of a keystore.
func ChangePassword(keystorePath string, oldPassword string, newPassword string) (Result, error) {
	if keystorePath == "" || newPassword == "" {
		return Result{Status: "failed", Error: "keystore_path and new_password are required"}, fmt.Errorf("keystore_path and new_password are required")
	}
	args := []string{
		"-storepasswd",
		"-keystore", keystorePath,
		"-storepass", oldPassword,
		"-new", newPassword,
	}
	out, err := runKeytool(args...)
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("change password: %v: %s", err, out)}, err
	}
	return Result{Status: "success", Changed: true, Message: out}, nil
}

// parseKeystoreList parses keytool -list -v output into CertInfo entries.
func parseKeystoreList(output string) []CertInfo {
	var certs []CertInfo
	lines := strings.Split(output, "\n")
	var current CertInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Alias name:") {
			if current.Alias != "" {
				certs = append(certs, current)
			}
			current = CertInfo{Valid: true}
			current.Alias = strings.TrimPrefix(line, "Alias name:")
			current.Alias = strings.TrimSpace(current.Alias)
		} else if strings.HasPrefix(line, "Entry type:") {
			current.Type = strings.TrimPrefix(line, "Entry type:")
			current.Type = strings.TrimSpace(current.Type)
		} else if strings.HasPrefix(line, "SHA") || strings.Contains(line, "Fingerprint") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				current.Fingerprint = parts[1]
			}
		}
	}
	if current.Alias != "" {
		certs = append(certs, current)
	}
	return certs
}
