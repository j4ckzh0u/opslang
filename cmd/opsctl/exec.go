// exec command for opsctl - remote execution via SSH
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute instructions on remote hosts via SSH",
	Long:  `Upload ops-runner to target hosts and execute JSON instruction packages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement exec logic
		return fmt.Errorf("exec command not yet implemented")
	},
}

func init() {
	execCmd.Flags().StringSlice("hosts", nil, "Target hosts (user@host)")
	execCmd.Flags().StringP("user", "u", "root", "SSH user")
	execCmd.Flags().StringP("key", "i", "", "SSH private key path")
	execCmd.Flags().StringP("password", "p", "", "SSH password")
	execCmd.Flags().String("inventory", "", "Inventory file path")
	execCmd.Flags().String("instructions", "", "JSON instructions file path")
}

// runExec is the main execution logic - placeholder
func runExec() error {
	fmt.Fprintln(os.Stderr, "exec: not yet implemented")
	return fmt.Errorf("not implemented")
}
