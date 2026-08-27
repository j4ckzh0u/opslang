//go:build opssec

// Package security implements permission checks, audit logging, resource limits, and signature verification.
package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// newHash returns a new SHA256 hash.
func newHash() hash.Hash {
	return sha256.New()
}

// SignatureManager handles Ed25519 signature operations.
type SignatureManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewSignatureManager creates a new signature manager.
// If privateKey is nil, a new key pair is generated.
func NewSignatureManager(privateKey ed25519.PrivateKey) *SignatureManager {
	if privateKey == nil {
		pub, priv, _ := ed25519.GenerateKey(rand.Reader)
		return &SignatureManager{
			privateKey: priv,
			publicKey:  pub,
		}
	}
	return &SignatureManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}
}

// Sign signs data and returns the signature.
func (m *SignatureManager) Sign(data []byte) []byte {
	return ed25519.Sign(m.privateKey, data)
}

// Verify verifies a signature against data.
func (m *SignatureManager) Verify(data, signature []byte) bool {
	return ed25519.Verify(m.publicKey, data, signature)
}

// PublicKey returns the public key.
func (m *SignatureManager) PublicKey() ed25519.PublicKey {
	return m.publicKey
}

// PrivateKey returns the private key.
func (m *SignatureManager) PrivateKey() ed25519.PrivateKey {
	return m.privateKey
}

// SignFile signs a file and returns the signature.
func (m *SignatureManager) SignFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return m.Sign(data), nil
}

// VerifyFile verifies a file's signature.
func (m *SignatureManager) VerifyFile(path string, signature []byte) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	return m.Verify(data, signature), nil
}

// SavePublicKey saves the public key to a file.
func (m *SignatureManager) SavePublicKey(path string) error {
	return os.WriteFile(path, m.publicKey, 0644)
}

// LoadPublicKey loads a public key from a file.
func LoadPublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}
	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(data))
	}
	return ed25519.PublicKey(data), nil
}

// SavePrivateKey saves the private key to a file.
func (m *SignatureManager) SavePrivateKey(path string) error {
	return os.WriteFile(path, m.privateKey, 0600)
}

// LoadPrivateKey loads a private key from a file.
func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	if len(data) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(data))
	}
	return ed25519.PrivateKey(data), nil
}

// SignatureToString converts a signature to hex string.
func SignatureToString(sig []byte) string {
	return hex.EncodeToString(sig)
}

// StringToSignature converts a hex string to signature.
func StringToSignature(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// PublicKeyToString converts a public key to hex string.
func PublicKeyToString(pub ed25519.PublicKey) string {
	return hex.EncodeToString(pub)
}

// StringToPublicKey converts a hex string to public key.
func StringToPublicKey(s string) (ed25519.PublicKey, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}
	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}
	return ed25519.PublicKey(data), nil
}

// ComputeChecksum computes SHA256 checksum of a file.
func ComputeChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := newHash()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
