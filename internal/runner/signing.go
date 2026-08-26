// runner/signing.go: Ed25519 signing of instruction packages.
//
// The controller signs each package before sending it; ops-runner verifies
// against a trusted public key when enforcement is requested (--pubkey).
// The signed payload is the canonical JSON encoding of the package with
// the Signature field cleared, so any mutation of instructions, privilege,
// dry-run flag, or task id invalidates the signature.
package runner

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SignPackage signs pkg with priv and stores the hex-encoded signature in
// pkg.Signature. Signing overwrites any previous signature.
func SignPackage(pkg *InstructionPackage, priv ed25519.PrivateKey) error {
	if len(priv) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid ed25519 private key size %d, want %d", len(priv), ed25519.PrivateKeySize)
	}
	payload, err := signaturePayload(pkg)
	if err != nil {
		return err
	}
	pkg.Signature = hex.EncodeToString(ed25519.Sign(priv, payload))
	return nil
}

// VerifyPackage checks pkg.Signature against pub. It returns an error for
// malformed inputs (missing signature, undecodable hex) and false for a
// well-formed signature that does not match — callers must treat both as
// rejection, but the distinction keeps diagnostics precise.
func VerifyPackage(pkg *InstructionPackage, pub ed25519.PublicKey) (bool, error) {
	if len(pub) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid ed25519 public key size %d, want %d", len(pub), ed25519.PublicKeySize)
	}
	if pkg.Signature == "" {
		return false, fmt.Errorf("instruction package carries no signature")
	}
	sig, err := hex.DecodeString(pkg.Signature)
	if err != nil {
		return false, fmt.Errorf("signature is not valid hex: %w", err)
	}
	payload, err := signaturePayload(pkg)
	if err != nil {
		return false, err
	}
	return ed25519.Verify(pub, payload, sig), nil
}

// signaturePayload marshals the package with the signature cleared. The
// struct is shallow-copied so verification never mutates the caller's
// package.
func signaturePayload(pkg *InstructionPackage) ([]byte, error) {
	clone := *pkg
	clone.Signature = ""
	data, err := json.Marshal(&clone)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package for signing: %w", err)
	}
	return data, nil
}
