// Startup wiring for controller-side SDK capabilities.
package main

import "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/file"

func init() {
	// file.distribute/file.collect need real SSH/SFTP transport on the
	// controller; without this they would fail with "no transfer
	// function configured" at runtime.
	file.WireSSHTransfer()
}
