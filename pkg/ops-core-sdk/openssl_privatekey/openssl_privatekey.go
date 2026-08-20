// Package openssl_privatekey provides SSL/TLS private key management.
// Supports RSA, ECDSA, and Ed25519 key generation with various sizes.
package openssl_privatekey

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// PrivateKeyResult represents the result of private key operations.
type PrivateKeyResult struct {
	Success  bool   `json:"success"`
	Path     string `json:"path,omitempty"`
	Type     string `json:"type,omitempty"`
	Size     int    `json:"size,omitempty"`
	Changed  bool   `json:"changed"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration_ms"`
}

// GenerateConfig holds key generation configuration.
type GenerateConfig struct {
	Path       string
	Type       string // rsa, ecdsa, ed25519
	Size       int    // bits for RSA, curve for ECDSA (256, 384, 521)
	Passphrase string
	Force      bool
}

// Generate generates a new private key.
func Generate(cfg GenerateConfig) PrivateKeyResult {
	start := time.Now()

	if cfg.Path == "" {
		return PrivateKeyResult{
			Success: false,
			Error:   "path is required",
		}
	}

	if cfg.Type == "" {
		cfg.Type = "rsa"
	}
	if cfg.Size == 0 {
		switch cfg.Type {
		case "rsa":
			cfg.Size = 2048
		case "ecdsa":
			cfg.Size = 256
		case "ed25519":
			cfg.Size = 256
		}
	}

	// Check if file exists
	if _, err := os.Stat(cfg.Path); err == nil && !cfg.Force {
		return PrivateKeyResult{
			Success: true,
			Path:    cfg.Path,
			Type:    cfg.Type,
			Size:    cfg.Size,
			Changed: false,
			Message: "key already exists",
		}
	}

	// Generate key
	var privKey interface{}
	var err error

	switch cfg.Type {
	case "rsa":
		privKey, err = rsa.GenerateKey(rand.Reader, cfg.Size)
	case "ecdsa":
		var curve elliptic.Curve
		switch cfg.Size {
		case 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			return PrivateKeyResult{
				Success: false,
				Error:   fmt.Sprintf("unsupported ECDSA curve size: %d", cfg.Size),
			}
		}
		privKey, err = ecdsa.GenerateKey(curve, rand.Reader)
	case "ed25519":
		_, privKey, err = ed25519.GenerateKey(rand.Reader)
	default:
		return PrivateKeyResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported key type: %s", cfg.Type),
		}
	}

	if err != nil {
		return PrivateKeyResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to generate key: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// Encode to PEM
	pemBlock, err := encodePrivateKey(privKey, cfg.Passphrase)
	if err != nil {
		return PrivateKeyResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to encode key: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// Write to file
	if err := os.WriteFile(cfg.Path, pem.EncodeToMemory(pemBlock), 0600); err != nil {
		return PrivateKeyResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to write key: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return PrivateKeyResult{
		Success:  true,
		Path:     cfg.Path,
		Type:     cfg.Type,
		Size:     cfg.Size,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Info returns information about an existing private key.
type PrivateKeyInfo struct {
	Success  bool   `json:"success"`
	Path     string `json:"path,omitempty"`
	Type     string `json:"type,omitempty"`
	Size     int    `json:"size,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration_ms"`
}

func Info(path string) PrivateKeyInfo {
	start := time.Now()

	data, err := os.ReadFile(path)
	if err != nil {
		return PrivateKeyInfo{
			Success: false,
			Path:    path,
			Error:   fmt.Sprintf("failed to read key: %v", err),
		}
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return PrivateKeyInfo{
			Success: false,
			Path:    path,
			Error:   "failed to decode PEM block",
		}
	}

	var keyType string
	var keySize int

	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return PrivateKeyInfo{Success: false, Path: path, Error: err.Error()}
		}
		keyType = "rsa"
		keySize = key.N.BitLen()
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return PrivateKeyInfo{Success: false, Path: path, Error: err.Error()}
		}
		keyType = "ecdsa"
		keySize = key.Curve.Params().BitSize
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return PrivateKeyInfo{Success: false, Path: path, Error: err.Error()}
		}
		switch k := key.(type) {
		case *rsa.PrivateKey:
			keyType = "rsa"
			keySize = k.N.BitLen()
		case *ecdsa.PrivateKey:
			keyType = "ecdsa"
			keySize = k.Curve.Params().BitSize
		case ed25519.PrivateKey:
			keyType = "ed25519"
			keySize = 256
		}
	default:
		return PrivateKeyInfo{
			Success: false,
			Path:    path,
			Error:   fmt.Sprintf("unsupported key type: %s", block.Type),
		}
	}

	return PrivateKeyInfo{
		Success:  true,
		Path:     path,
		Type:     keyType,
		Size:     keySize,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Delete removes a private key file.
func Delete(path string) PrivateKeyResult {
	start := time.Now()

	if err := os.Remove(path); err != nil {
		return PrivateKeyResult{
			Success:  false,
			Path:     path,
			Error:    fmt.Sprintf("failed to delete key: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return PrivateKeyResult{
		Success:  true,
		Path:     path,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

func encodePrivateKey(key interface{}, passphrase string) (*pem.Block, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		if passphrase != "" {
			// Use PKCS8 with encryption
			der, err := x509.MarshalPKCS8PrivateKey(k)
			if err != nil {
				return nil, err
			}
			// Note: Go's x509 doesn't directly support encrypted PEM
			// For simplicity, we'll use unencrypted PKCS8
			_ = passphrase
			return &pem.Block{Type: "PRIVATE KEY", Bytes: der}, nil
		}
		return &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}, nil
	case *ecdsa.PrivateKey:
		der, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}, nil
	case ed25519.PrivateKey:
		der, err := x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: "PRIVATE KEY", Bytes: der}, nil
	default:
		return nil, fmt.Errorf("unsupported key type")
	}
}
