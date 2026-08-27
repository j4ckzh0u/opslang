// ops-runner is the universal runner that executes JSON instruction packages
// from stdin and outputs JSON results to stdout.
//
// Exit codes:
//
//	0 - all instructions succeeded (status "ok")
//	1 - some instructions failed (status "partial")
//	2 - every instruction failed (status "failed")
//	3 - protocol/usage error (bad input, unsupported version)
package main

import (
	"crypto/ed25519"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/j4ckzh0u/opslang/internal/runner"
)

var (
	dryRun  bool
	version bool
	// pubKeyPath points at the trusted Ed25519 public key (raw bytes, as
	// written by security.SavePublicKey / opsctl keygen). When set, the
	// runner refuses unsigned or tampered instruction packages.
	pubKeyPath string
)

func main() {
	flag.BoolVar(&dryRun, "dry-run", false, "Execute in dry-run mode (no actual changes)")
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.StringVar(&pubKeyPath, "pubkey", "", "Ed25519 public key file; when set, unsigned or tampered packages are refused")
	flag.Parse()

	if version {
		fmt.Println("ops-runner v0.1.0")
		os.Exit(0)
	}

	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(3)
	}

	// The run itself may have partially or fully failed; report it via the
	// exit code so the controller (and CI) can detect it without parsing JSON.
	switch status {
	case "partial":
		os.Exit(1)
	case "failed":
		os.Exit(2)
	}
}

// status records the outcome of the last run for exit-code mapping.
var status string

func run(in io.Reader, out io.Writer) error {
	// Read instruction package from stdin.
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Parse instruction package.
	var pkg runner.InstructionPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("failed to parse instruction package: %w", err)
	}

	// Validate version.
	if pkg.Version != "1.0" {
		return fmt.Errorf("unsupported protocol version: %s (expected 1.0)", pkg.Version)
	}

	// Signature enforcement runs before anything else touches the package:
	// a tampered package must never execute, not even in dry-run mode.
	if pubKeyPath != "" {
		pub, lerr := loadTrustedPublicKey()
		if lerr != nil {
			// A missing/corrupt key file is an operator error, not a
			// package failure: report as usage error (exit 3).
			return lerr
		}
		ok, verr := runner.VerifyPackage(&pkg, pub)
		if verr != nil || !ok {
			// Anything wrong with the PACKAGE's signature — missing,
			// undecodable, or mismatched — is a security rejection with
			// structured output so the controller can parse it.
			return emitSecurityRejection(out, verr)
		}
	}

	// Apply dry-run flag.
	if dryRun {
		pkg.DryRun = true
	}

	// Create registry with all standard operations.
	registry := runner.NewRegistry()

	// Execute instructions.
	output := runner.Run(&pkg, registry)
	status = output.Status

	// Marshal output.
	result, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	// Write output.
	if _, err := out.Write(result); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	if _, err := out.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// loadTrustedPublicKey reads the enforcement public key, accepting both
// raw key bytes (as written by security.SavePublicKey / opsctl keygen) and
// hex encoding.
func loadTrustedPublicKey() (ed25519.PublicKey, error) {
	keyData, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key %s: %w", pubKeyPath, err)
	}
	if pub, herr := parseHexPublicKey(strings.TrimSpace(string(keyData))); herr == nil {
		return pub, nil
	}
	if len(keyData) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key file %s is neither %d-byte raw nor valid hex", pubKeyPath, ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(keyData), nil
}

// emitSecurityRejection reports a tampered or unsigned package as a
// structured failure on stdout (so the controller's parser sees it) plus
// the "failed" status for exit-code mapping, instead of a bare usage
// error — the distinction matters: the package arrived but must not run.
func emitSecurityRejection(out io.Writer, reason error) error {
	status = "failed"
	msg := "security: instruction package signature is missing or invalid; refusing execution"
	if reason != nil {
		msg = fmt.Sprintf("security: refusing instruction package: %v", reason)
	}
	rejection := runner.Output{
		Status: "failed",
		Errors: []string{msg},
	}
	result, err := json.MarshalIndent(rejection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rejection output: %w", err)
	}
	if _, werr := out.Write(append(result, '\n')); werr != nil {
		return fmt.Errorf("failed to write rejection output: %w", werr)
	}
	fmt.Fprintln(os.Stderr, "error:", msg)
	return nil
}
