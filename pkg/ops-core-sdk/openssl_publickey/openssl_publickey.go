// Package openssl_publickey provides public key extraction from private keys.
// Supports RSA, ECDSA, and Ed25519 private keys.
package openssl_publickey

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// PublicKeyResult represents the result of public key operations.
type PublicKeyResult struct {
	Success    bool   `json:"success"`
	OutputFile string `json:"output_file,omitempty"`
	Changed    bool   `json:"changed,omitempty"`
	Error      string `json:"error,omitempty"`
	Duration   int64  `json:"duration_ms"`
}

// Extract extracts the public key from a private key file.
func Extract(privateKeyFile, outputFile string, force bool) PublicKeyResult {
	start := time.Now()

	if privateKeyFile == "" {
		return PublicKeyResult{
			Success: false,
			Error:   "private_key_file is required",
		}
	}

	if outputFile == "" {
		return PublicKeyResult{
			Success: false,
			Error:   "output_file is required",
		}
	}

	// Check if output file exists
	if !force {
		if _, err := os.Stat(outputFile); err == nil {
			return PublicKeyResult{
				Success: true,
				Changed: false,
			}
		}
	}

	// Read private key
	keyData, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return PublicKeyResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read private key: %v", err),
		}
	}

	// Parse PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return PublicKeyResult{
			Success: false,
			Error:   "failed to decode PEM block",
		}
	}

	// Parse private key and extract public key
	var publicKey interface{}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		switch k := key.(type) {
		case *rsa.PrivateKey:
			publicKey = &k.PublicKey
		case *ecdsa.PrivateKey:
			publicKey = &k.PublicKey
		case ed25519.PrivateKey:
			publicKey = k.Public()
		default:
			return PublicKeyResult{
				Success: false,
				Error:   "unsupported key type in PKCS8",
			}
		}
	} else if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		publicKey = &key.PublicKey
	} else if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		publicKey = &key.PublicKey
	} else {
		return PublicKeyResult{
			Success: false,
			Error:   "failed to parse private key",
		}
	}

	// Marshal public key to PEM
	var pubKeyBytes []byte
	var blockType string

	switch k := publicKey.(type) {
	case *rsa.PublicKey:
		pubKeyBytes, err = x509.MarshalPKIXPublicKey(k)
		blockType = "PUBLIC KEY"
	case *ecdsa.PublicKey:
		pubKeyBytes, err = x509.MarshalPKIXPublicKey(k)
		blockType = "PUBLIC KEY"
	case ed25519.PublicKey:
		pubKeyBytes, err = x509.MarshalPKIXPublicKey(k)
		blockType = "PUBLIC KEY"
	default:
		return PublicKeyResult{
			Success: false,
			Error:   "unsupported public key type",
		}
	}

	if err != nil {
		return PublicKeyResult{
			Success: false,
			Error:   fmt.Sprintf("failed to marshal public key: %v", err),
		}
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  blockType,
		Bytes: pubKeyBytes,
	})

	// Write to file
	if err := os.WriteFile(outputFile, pubKeyPEM, 0644); err != nil {
		return PublicKeyResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write public key: %v", err),
		}
	}

	return PublicKeyResult{
		Success:    true,
		OutputFile: outputFile,
		Changed:    true,
		Duration:   time.Since(start).Milliseconds(),
	}
}

// Info returns information about a public key file.
func Info(publicKeyFile string) PublicKeyResult {
	start := time.Now()

	data, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return PublicKeyResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read public key: %v", err),
		}
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return PublicKeyResult{
			Success: false,
			Error:   "failed to decode PEM block",
		}
	}

	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return PublicKeyResult{
			Success: false,
			Error:   fmt.Sprintf("failed to parse public key: %v", err),
		}
	}

	return PublicKeyResult{
		Success:    true,
		OutputFile: publicKeyFile,
		Duration:   time.Since(start).Milliseconds(),
	}
}

// Delete removes a public key file.
func Delete(publicKeyFile string) PublicKeyResult {
	start := time.Now()

	// Check if file exists
	if _, err := os.Stat(publicKeyFile); os.IsNotExist(err) {
		return PublicKeyResult{
			Success: true,
			Changed: false,
		}
	}

	if err := os.Remove(publicKeyFile); err != nil {
		return PublicKeyResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to delete public key: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return PublicKeyResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// GetKeyType returns the type of a public key (RSA, ECDSA, Ed25519).
func GetKeyType(publicKeyFile string) (string, error) {
	data, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read public key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	switch k := pub.(type) {
	case *rsa.PublicKey:
		return "RSA", nil
	case *ecdsa.PublicKey:
		switch k.Curve {
		case elliptic.P256():
			return "ECDSA-P256", nil
		case elliptic.P384():
			return "ECDSA-P384", nil
		case elliptic.P521():
			return "ECDSA-P521", nil
		default:
			return "ECDSA", nil
		}
	case ed25519.PublicKey:
		return "Ed25519", nil
	default:
		return "Unknown", nil
	}
}
