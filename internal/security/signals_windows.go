//go:build windows

package security

func ensureGlobalSignalHandler() {
	// Windows process termination is handled by the operating system. Cleanup
	// remains available through the explicit Cleanup method and deferred calls.
}
