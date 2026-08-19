// Package synchronize provides file synchronization operations (rsync wrapper).
package synchronize

import (
	"fmt"
	"os/exec"
)

// Result represents the result of a sync operation.
type Result struct {
	Source    string `json:"source"`
	Dest      string `json:"dest"`
	Changed   bool   `json:"changed"`
	Message   string `json:"message"`
}

// Sync synchronizes files from source to destination using rsync.
func Sync(source string, dest string, deleteExtra bool, compress bool) (*Result, error) {
	args := []string{"-a"}
	if compress {
		args = append(args, "-z")
	}
	if deleteExtra {
		args = append(args, "--delete")
	}
	args = append(args, source, dest)

	cmd := exec.Command("rsync", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("rsync failed: %w (output: %s)", err, string(out))
	}
	return &Result{
		Source:  source,
		Dest:    dest,
		Changed: true,
		Message: "Sync completed",
	}, nil
}
