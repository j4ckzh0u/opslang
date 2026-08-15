package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// Sign signs the payload with the private key using Ed25519
func Sign(payload []byte, privKey ed25519.PrivateKey) []byte {
	return ed25519.Sign(privKey, payload)
}

// Verify verifies the signature of the payload with the public key
func Verify(payload, sig []byte, pubKey ed25519.PublicKey) bool {
	return ed25519.Verify(pubKey, payload, sig)
}

// GenerateKey generates a new Ed25519 key pair
func GenerateKey() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// SavePublicKey saves the public key to a file in PEM format
func SavePublicKey(path string, key ed25519.PublicKey) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal the public key to PKIX format
	derBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}

	// Write to file
	if err := os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

// LoadPublicKey loads a public key from a PEM file
func LoadPublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	// Parse the public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	edKey, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not Ed25519")
	}

	return edKey, nil
}

// SavePrivateKey saves the private key to a file in PEM format
func SavePrivateKey(path string, key ed25519.PrivateKey) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal the private key to PKCS8 format
	derBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: derBytes,
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

// LoadPrivateKey loads a private key from a PEM file
func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	// Parse the private key
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	edKey, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not Ed25519")
	}

	return edKey, nil
}
