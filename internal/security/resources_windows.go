//go:build windows

package security

func applyPlatformResourceLimits(_ *ResourceLimits) error {
	// Windows resource jobs require a process-wide supervisor; keep this API
	// side-effect free until that supervisor is available.
	return nil
}
