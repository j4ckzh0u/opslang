//go:build !opssec

package main

import (
	"crypto/ed25519"
	"errors"
)

// errHexParseDisabled reports that hex key parsing (opsctl keygen format)
// is unavailable: this binary was built without -tags opssec, so the
// signing half of internal/security is not compiled in. Raw 32-byte
// public-key files keep working via the stdlib path in main.
var errHexParseDisabled = errors.New("hex public key parsing requires an opssec build; use a raw 32-byte key file instead")

func parseHexPublicKey(string) (ed25519.PublicKey, error) {
	return nil, errHexParseDisabled
}
