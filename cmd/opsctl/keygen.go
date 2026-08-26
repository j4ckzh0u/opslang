// keygen generates an Ed25519 signing key pair for instruction packages.
// The private key stays on the controller (--sign-key); the public key is
// distributed to target hosts and enforced by ops-runner --pubkey.
package main

import (
	"fmt"
	"os"

	"github.com/j4ckzh0u/opslang/internal/security"
	"github.com/spf13/cobra"
)

var (
	keygenOut   string
	keygenForce bool
)

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate an Ed25519 signing key pair for instruction packages",
	Long: `Generate an Ed25519 signing key pair.

The private key signs deploy instruction packages (opsctl deploy/exec
--sign-key). The public key is placed on target hosts at a fixed path and
enforced by ops-runner via its --pubkey flag.

Existing files are never overwritten unless --force is given.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runKeygen()
	},
}

func init() {
	keygenCmd.Flags().StringVarP(&keygenOut, "out", "o", "opslang-signing", "Output prefix; writes <prefix>.key and <prefix>.pub")
	keygenCmd.Flags().BoolVar(&keygenForce, "force", false, "Overwrite existing key files")
}

func runKeygen() error {
	if keygenOut == "" {
		return fmt.Errorf("--out prefix must not be empty")
	}
	privPath := keygenOut + ".key"
	pubPath := keygenOut + ".pub"

	for _, p := range []string{privPath, pubPath} {
		if _, err := os.Stat(p); err == nil && !keygenForce {
			return fmt.Errorf("%s already exists (use --force to overwrite)", p)
		}
	}

	mgr := security.NewSignatureManager(nil)
	if err := mgr.SavePrivateKey(privPath); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}
	if err := mgr.SavePublicKey(pubPath); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	fmt.Printf("Private key written to %s (mode 0600)\n", privPath)
	fmt.Printf("Public key written to  %s (mode 0644)\n", pubPath)
	fmt.Println("\nUsage:")
	fmt.Printf("  sign:      opsctl deploy script.ops --sign-key %s\n", privPath)
	fmt.Printf("  enforce:   distribute %s to targets, then deploy with --verify-key <remote-path>\n", pubPath)
	return nil
}
