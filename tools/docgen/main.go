// Command docgen regenerates docs/generated/ops-index.md from opsspec.
// Usage: go run ./tools/docgen [output-path]
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/j4ckzh0u/opslang/internal/docgen"
)

func main() {
	out := "docs/generated/ops-index.md"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	content, err := docgen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "docgen: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "docgen: create dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "docgen: write %s: %v\n", out, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d ops)\n", out, docgen.OpCount())
}
