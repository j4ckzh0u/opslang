// opsctl is the main CLI for OpsLang.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "opsctl",
	Short: "OpsLang control CLI",
	Long:  "opsctl is the command-line interface for OpsLang, a domain-specific language for operations.",
}

func init() {
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(replCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(keygenCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of opsctl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("opsctl v0.1.0")
	},
}
