//go:build opssec

package main

import (
	"crypto/ed25519"

	"github.com/j4ckzh0u/opslang/internal/security"
)

// parseHexPublicKey decodes the hex-encoded form written by opsctl keygen.
// Compiled only under -tags opssec (the side-effect half of internal/security).
func parseHexPublicKey(s string) (ed25519.PublicKey, error) {
	pub, herr := security.StringToPublicKey(s)
	if herr != nil {
		return nil, herr
	}
	return pub, nil
}
