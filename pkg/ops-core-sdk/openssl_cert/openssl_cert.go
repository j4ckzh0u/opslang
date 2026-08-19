// Package openssl_cert provides SSL/TLS certificate and CSR management.
// Uses exec.Command to invoke openssl binary.
package openssl_cert

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CertInfo represents certificate metadata.
type CertInfo struct {
	Subject    string `json:"subject"`
	Issuer     string `json:"issuer"`
	NotBefore  string `json:"not_before"`
	NotAfter   string `json:"not_after"`
	Serial     string `json:"serial"`
	SelfSigned bool   `json:"self_signed"`
}

// CSRResult is returned by CreateCSR.
type CSRResult struct {
	CSRPath  string `json:"csr_path"`
	KeyPath  string `json:"key_path"`
	Success  bool   `json:"success"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// SelfSignedResult is returned by GenerateSelfSigned.
type SelfSignedResult struct {
	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
	Success  bool   `json:"success"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// InspectResult is returned by Inspect.
type InspectResult struct {
	Exists bool     `json:"exists"`
	Info   CertInfo `json:"info,omitempty"`
	Error  string   `json:"error,omitempty"`
}

// VerifyResult is returned by Verify.
type VerifyResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
	Output string   `json:"output,omitempty"`
}

// ActionResult is returned by mutating operations.
type ActionResult struct {
	Success  bool   `json:"success"`
	Changed  bool   `json:"changed"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// ExpiryResult is returned by CheckExpiry.
type ExpiryResult struct {
	Path        string `json:"path"`
	ExpiresAt   string `json:"expires_at"`
	DaysLeft    int    `json:"days_left"`
	Expired     bool   `json:"expired"`
}

func opensslCmd(args ...string) *exec.Cmd {
	return exec.Command("openssl", args...)
}

// CreateCSR generates a certificate signing request.
func CreateCSR(keyPath, csrPath, subject string, keyBits int) (CSRResult, error) {
	start := time.Now()
	if keyPath == "" || csrPath == "" || subject == "" {
		return CSRResult{}, fmt.Errorf("key_path, csr_path, and subject are required")
	}
	if keyBits <= 0 {
		keyBits = 2048
	}

	// Generate private key
	cmd := opensslCmd("genrsa", "-out", keyPath, fmt.Sprintf("%d", keyBits))
	if out, err := cmd.CombinedOutput(); err != nil {
		return CSRResult{Error: string(out), Duration: time.Since(start).Milliseconds()}, err
	}

	// Generate CSR
	cmd = opensslCmd("req", "-new", "-key", keyPath, "-out", csrPath, "-subj", subject)
	if out, err := cmd.CombinedOutput(); err != nil {
		return CSRResult{Error: string(out), Duration: time.Since(start).Milliseconds()}, err
	}

	return CSRResult{CSRPath: csrPath, KeyPath: keyPath, Success: true, Duration: time.Since(start).Milliseconds()}, nil
}

// GenerateSelfSigned generates a self-signed certificate.
func GenerateSelfSigned(certPath, keyPath, subject string, days int, keyBits int) (SelfSignedResult, error) {
	start := time.Now()
	if certPath == "" || keyPath == "" || subject == "" {
		return SelfSignedResult{}, fmt.Errorf("cert_path, key_path, and subject are required")
	}
	if days <= 0 {
		days = 365
	}
	if keyBits <= 0 {
		keyBits = 2048
	}

	cmd := opensslCmd("req", "-x509", "-newkey", fmt.Sprintf("rsa:%d", keyBits),
		"-keyout", keyPath, "-out", certPath, "-days", fmt.Sprintf("%d", days),
		"-nodes", "-subj", subject)
	if out, err := cmd.CombinedOutput(); err != nil {
		return SelfSignedResult{Error: string(out), Duration: time.Since(start).Milliseconds()}, err
	}

	return SelfSignedResult{CertPath: certPath, KeyPath: keyPath, Success: true, Duration: time.Since(start).Milliseconds()}, nil
}

// Inspect returns metadata about a certificate.
func Inspect(certPath string) (InspectResult, error) {
	if certPath == "" {
		return InspectResult{}, fmt.Errorf("cert_path is required")
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return InspectResult{Exists: false}, nil
	}

	cmd := opensslCmd("x509", "-in", certPath, "-noout", "-subject", "-issuer", "-dates", "-serial")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return InspectResult{Exists: true, Error: string(out)}, nil
	}

	info := CertInfo{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "subject=") {
			info.Subject = strings.TrimPrefix(line, "subject=")
		} else if strings.HasPrefix(line, "issuer=") {
			info.Issuer = strings.TrimPrefix(line, "issuer=")
		} else if strings.HasPrefix(line, "notBefore=") {
			info.NotBefore = strings.TrimPrefix(line, "notBefore=")
		} else if strings.HasPrefix(line, "notAfter=") {
			info.NotAfter = strings.TrimPrefix(line, "notAfter=")
		} else if strings.HasPrefix(line, "serial=") {
			info.Serial = strings.TrimPrefix(line, "serial=")
		}
	}
	info.SelfSigned = info.Subject == info.Issuer

	return InspectResult{Exists: true, Info: info}, nil
}

// Verify checks if a certificate is valid.
func Verify(certPath, caPath string) (VerifyResult, error) {
	if certPath == "" {
		return VerifyResult{}, fmt.Errorf("cert_path is required")
	}
	args := []string{"verify"}
	if caPath != "" {
		args = append(args, "-CAfile", caPath)
	}
	args = append(args, certPath)

	cmd := opensslCmd(args...)
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return VerifyResult{Valid: false, Errors: []string{output}, Output: output}, nil
	}
	return VerifyResult{Valid: true, Output: output}, nil
}

// CheckExpiry checks how many days until a certificate expires.
func CheckExpiry(certPath string) (ExpiryResult, error) {
	if certPath == "" {
		return ExpiryResult{}, fmt.Errorf("cert_path is required")
	}

	cmd := opensslCmd("x509", "-in", certPath, "-noout", "-enddate")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ExpiryResult{Path: certPath}, fmt.Errorf("failed to read cert: %w", err)
	}

	line := strings.TrimSpace(string(out))
	dateStr := strings.TrimPrefix(line, "notAfter=")

	// Parse the date (openssl format: Mon DD HH:MM:SS YYYY GMT)
	t, err := time.Parse("Jan 2 15:04:05 2006 MST", dateStr)
	if err != nil {
		return ExpiryResult{Path: certPath, ExpiresAt: dateStr}, nil
	}

	daysLeft := int(time.Until(t).Hours() / 24)
	return ExpiryResult{
		Path:      certPath,
		ExpiresAt: t.Format(time.RFC3339),
		DaysLeft:  daysLeft,
		Expired:   daysLeft < 0,
	}, nil
}

// ConvertFormat converts a certificate between PEM and DER formats.
func ConvertFormat(inputPath, outputPath, outputFormat string) (ActionResult, error) {
	start := time.Now()
	if inputPath == "" || outputPath == "" {
		return ActionResult{}, fmt.Errorf("input_path and output_path are required")
	}
	if outputFormat == "" {
		outputFormat = "der"
	}

	var args []string
	if outputFormat == "der" {
		args = []string{"x509", "-in", inputPath, "-outform", "DER", "-out", outputPath}
	} else {
		args = []string{"x509", "-in", inputPath, "-inform", "DER", "-outform", "PEM", "-out", outputPath}
	}

	cmd := opensslCmd(args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return ActionResult{Error: string(out), Duration: time.Since(start).Milliseconds()}, err
	}
	return ActionResult{Success: true, Changed: true, Duration: time.Since(start).Milliseconds()}, nil
}
