package sys

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// UUIDResult represents the result of UUID generation.
type UUIDResult struct {
	UUID string `json:"uuid"`
}

// PasswordResult represents the result of password generation.
type PasswordResult struct {
	Password string `json:"password"`
	Length   int    `json:"length"`
}

// UUID generates a random UUID (v4).
func UUID() (UUIDResult, error) {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		return UUIDResult{}, fmt.Errorf("sys.UUID: failed to generate UUID: %w", err)
	}
	// Set version 4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Set variant 10
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return UUIDResult{
		UUID: fmt.Sprintf("%s-%s-%s-%s-%s",
			hex.EncodeToString(uuid[0:4]),
			hex.EncodeToString(uuid[4:6]),
			hex.EncodeToString(uuid[6:8]),
			hex.EncodeToString(uuid[8:10]),
			hex.EncodeToString(uuid[10:16])),
	}, nil
}

// RandomPassword generates a cryptographically secure random password.
// length specifies the password length (minimum 8).
// useSpecial, useNumbers, useUppercase control character set inclusion.
func RandomPassword(length int, useSpecial bool, useNumbers bool, useUppercase bool) (PasswordResult, error) {
	if length < 8 {
		length = 8
	}

	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	charset := lowercase
	if useUppercase {
		charset += uppercase
	}
	if useNumbers {
		charset += numbers
	}
	if useSpecial {
		charset += special
	}

	// Ensure at least one character from each selected set
	var password strings.Builder
	if useUppercase {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(uppercase))))
		password.WriteByte(uppercase[idx.Int64()])
	}
	if useNumbers {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(numbers))))
		password.WriteByte(numbers[idx.Int64()])
	}
	if useSpecial {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(special))))
		password.WriteByte(special[idx.Int64()])
	}

	// Fill remaining length with random characters from the full charset
	remaining := length - password.Len()
	for i := 0; i < remaining; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password.WriteByte(charset[idx.Int64()])
	}

	// Shuffle the password to avoid predictable positions
	runes := []rune(password.String())
	for i := len(runes) - 1; i > 0; i-- {
		jBig, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(jBig.Int64())
		runes[i], runes[j] = runes[j], runes[i]
	}

	return PasswordResult{
		Password: string(runes),
		Length:   len(runes),
	}, nil
}
