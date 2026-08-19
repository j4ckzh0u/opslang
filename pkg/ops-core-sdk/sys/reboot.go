// System reboot.
package sys

import (
	"fmt"
	"os/exec"
)

// RebootResult is returned by Reboot.
type RebootResult struct {
	Message string `json:"message"`
}

// Reboot initiates a system reboot.
func Reboot() (RebootResult, error) {
	cmd := exec.Command("shutdown", "-r", "now")
	if err := cmd.Run(); err != nil {
		// Fallback to reboot command
		cmd2 := exec.Command("reboot")
		if err2 := cmd2.Run(); err2 != nil {
			return RebootResult{}, fmt.Errorf("sys.Reboot: %w (fallback: %v)", err, err2)
		}
	}
	return RebootResult{Message: "reboot initiated"}, nil
}
