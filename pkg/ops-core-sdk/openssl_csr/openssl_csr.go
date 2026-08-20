// Package openssl_csr provides Certificate Signing Request generation.
// Supports generating CSRs from existing private keys for SSL/TLS certificates.
package openssl_csr

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// CSRConfig holds CSR configuration.
type CSRConfig struct {
	CommonName         string
	Organization       string
	OrganizationalUnit string
	Country            string
	State              string
	Locality           string
	Email              string
	DNSNames           []string
	IPAddresses        []string
	KeyFile            string
	OutputFile         string
	Force              bool
}

// CSRResult represents the result of CSR generation.
type CSRResult struct {
	Success    bool   `json:"success"`
	OutputFile string `json:"output_file,omitempty"`
	Changed    bool   `json:"changed,omitempty"`
	Error      string `json:"error,omitempty"`
	Duration   int64  `json:"duration_ms"`
}

// Generate creates a Certificate Signing Request.
func Generate(cfg CSRConfig) CSRResult {
	start := time.Now()

	if cfg.CommonName == "" {
		return CSRResult{
			Success: false,
			Error:   "common_name is required",
		}
	}

	if cfg.KeyFile == "" {
		return CSRResult{
			Success: false,
			Error:   "key_file is required",
		}
	}

	if cfg.OutputFile == "" {
		return CSRResult{
			Success: false,
			Error:   "output_file is required",
		}
	}

	// Check if output file exists
	if !cfg.Force {
		if _, err := os.Stat(cfg.OutputFile); err == nil {
			return CSRResult{
				Success: true,
				Changed: false,
			}
		}
	}

	// Read private key
	keyData, err := os.ReadFile(cfg.KeyFile)
	if err != nil {
		return CSRResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read key file: %v", err),
		}
	}

	// Parse private key (try RSA first, then ECDSA, then Ed25519)
	block, _ := pem.Decode(keyData)
	if block == nil {
		return CSRResult{
			Success: false,
			Error:   "failed to decode PEM block",
		}
	}

	var privateKey interface{}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		privateKey = key
	} else if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		privateKey = key
	} else if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		privateKey = key
	} else {
		return CSRResult{
			Success: false,
			Error:   "failed to parse private key",
		}
	}

	// Build subject
	subject := pkix.Name{
		CommonName: cfg.CommonName,
	}
	if cfg.Organization != "" {
		subject.Organization = []string{cfg.Organization}
	}
	if cfg.OrganizationalUnit != "" {
		subject.OrganizationalUnit = []string{cfg.OrganizationalUnit}
	}
	if cfg.Country != "" {
		subject.Country = []string{cfg.Country}
	}
	if cfg.State != "" {
		subject.Province = []string{cfg.State}
	}
	if cfg.Locality != "" {
		subject.Locality = []string{cfg.Locality}
	}

	// Create CSR template
	template := &x509.CertificateRequest{
		Subject:        subject,
		DNSNames:       cfg.DNSNames,
		EmailAddresses: []string{cfg.Email},
	}

	// Set signature algorithm based on key type
	switch k := privateKey.(type) {
	case *ecdsa.PrivateKey:
		switch k.Curve {
		case elliptic.P256():
			template.SignatureAlgorithm = x509.ECDSAWithSHA256
		case elliptic.P384():
			template.SignatureAlgorithm = x509.ECDSAWithSHA384
		case elliptic.P521():
			template.SignatureAlgorithm = x509.ECDSAWithSHA512
		default:
			template.SignatureAlgorithm = x509.ECDSAWithSHA256
		}
	case *rsa.PrivateKey:
		template.SignatureAlgorithm = x509.SHA256WithRSA
	case ed25519.PrivateKey:
		template.SignatureAlgorithm = x509.PureEd25519
	default:
		template.SignatureAlgorithm = x509.SHA256WithRSA
	}

	// Generate CSR
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, template, privateKey)
	if err != nil {
		return CSRResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create CSR: %v", err),
		}
	}

	// Encode to PEM
	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	// Write to file
	if err := os.WriteFile(cfg.OutputFile, csrPEM, 0644); err != nil {
		return CSRResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write CSR: %v", err),
		}
	}

	return CSRResult{
		Success:    true,
		OutputFile: cfg.OutputFile,
		Changed:    true,
		Duration:   time.Since(start).Milliseconds(),
	}
}

// Info returns information about an existing CSR.
func Info(csrFile string) CSRResult {
	start := time.Now()

	data, err := os.ReadFile(csrFile)
	if err != nil {
		return CSRResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read CSR: %v", err),
		}
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return CSRResult{
			Success: false,
			Error:   "failed to decode PEM block",
		}
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return CSRResult{
			Success: false,
			Error:   fmt.Sprintf("failed to parse CSR: %v", err),
		}
	}

	// Verify CSR signature
	if err := csr.CheckSignature(); err != nil {
		return CSRResult{
			Success: false,
			Error:   fmt.Sprintf("invalid CSR signature: %v", err),
		}
	}

	return CSRResult{
		Success:    true,
		OutputFile: csrFile,
		Duration:   time.Since(start).Milliseconds(),
	}
}

// Delete removes a CSR file.
func Delete(csrFile string) CSRResult {
	start := time.Now()

	// Check if file exists
	if _, err := os.Stat(csrFile); os.IsNotExist(err) {
		return CSRResult{
			Success: true,
			Changed: false,
		}
	}

	if err := os.Remove(csrFile); err != nil {
		return CSRResult{
			Success: false,
			Error:   fmt.Sprintf("failed to delete CSR: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return CSRResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}
