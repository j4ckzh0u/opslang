// Command genshow prints AOT-generated Go for a script (debug aid).
package main

import (
	"fmt"
	"os"

	"github.com/j4ckzh0u/opslang/internal/compiler"
)

func main() {
	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code, err := compiler.GenerateCode(string(src), os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
	os.Stdout.WriteString(code)
}
