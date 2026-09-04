//go:build !windows

package process

import "syscall"

var signals = map[string]syscall.Signal{
	"TERM": syscall.SIGTERM,
	"KILL": syscall.SIGKILL,
	"HUP":  syscall.SIGHUP,
	"INT":  syscall.SIGINT,
	"USR1": syscall.SIGUSR1,
	"USR2": syscall.SIGUSR2,
}

func supportedSignal(name string) bool {
	_, ok := signals[name]
	return ok
}

func killProcess(pid int, name string) error {
	return syscall.Kill(pid, signals[name])
}
