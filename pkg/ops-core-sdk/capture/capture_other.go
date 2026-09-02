//go:build !linux

// Package capture: passive packet capture. Linux-only by design (AF_PACKET);
// on every other platform this stub returns an explicit error so scripts
// fail loudly instead of silently, and CGO stays disabled everywhere.
package capture

import "errors"

var errUnsupported = errors.New("net.capture: passive capture requires Linux (AF_PACKET); rebuild the tool on a Linux host")

func Capture(Options) (*Result, error) { return nil, errUnsupported }

// Run mirrors the Linux signature so generated code compiles everywhere;
// execution returns errUnsupported.
func Run(_ string, _ int, _ int, _ string) (Result, error) {
	return Result{}, errUnsupported
}
