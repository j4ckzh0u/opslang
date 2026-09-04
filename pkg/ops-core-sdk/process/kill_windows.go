//go:build windows

package process

import (
	"fmt"
	"os"
)

func supportedSignal(name string) bool {
	switch name {
	case "TERM", "KILL", "HUP", "INT", "USR1", "USR2":
		return true
	default:
		return false
	}
}

func killProcess(pid int, name string) error {
	if name != "TERM" && name != "KILL" && name != "INT" {
		return fmt.Errorf("signal %s is unsupported on Windows; supported signals: TERM, KILL, INT", name)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}
