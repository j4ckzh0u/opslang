// build command for opsctl - AOT compile OpsLang scripts to static binaries
package main

import (
	"fmt"

	"github.com/opslang/opslang/internal/compiler"
	"github.com/spf13/cobra"
)

var (
	buildSource    string
	buildOutput    string
	buildTargetArch string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compile an OpsLang script to a static binary",
	Long: `Compile an OpsLang script into a standalone static binary using AOT compilation.

The compiled binary includes the ops-core-sdk and can run on the target platform
without any runtime dependencies. The compilation result is cached based on source
content and target architecture for fast subsequent builds.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuildCommand()
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildSource, "source", "s", "", "OpsLang source file path (required)")
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "./output", "Output binary path")
	buildCmd.Flags().StringVar(&buildTargetArch, "target-arch", "", "Target architecture (e.g., linux/amd64, linux/arm64). Default: current platform")
}

func runBuildCommand() error {
	if buildSource == "" {
		return fmt.Errorf("--source flag is required")
	}

	c, err := compiler.NewCompiler()
	if err != nil {
		return fmt.Errorf("failed to initialize compiler: %w", err)
	}

	fmt.Printf("Compiling %s -> %s\n", buildSource, buildOutput)
	if buildTargetArch != "" {
		fmt.Printf("Target: %s\n", buildTargetArch)
	}

	if err := c.Compile(buildSource, buildTargetArch, buildOutput); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	fmt.Println("Build successful!")
	return nil
}
