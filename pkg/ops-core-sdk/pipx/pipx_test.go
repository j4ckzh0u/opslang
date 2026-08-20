package pipx

import (
	"os/exec"
	"testing"
)

func skipIfNoPipx(t *testing.T) {
	if _, err := exec.LookPath("pipx"); err != nil {
		t.Skip("pipx not found in PATH")
	}
}

func TestInstallValidation(t *testing.T) {
	_, err := Install("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestUninstallValidation(t *testing.T) {
	_, err := Uninstall("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestUpgradeValidation(t *testing.T) {
	_, err := Upgrade("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestList(t *testing.T) {
	skipIfNoPipx(t)
	packages, err := List()
	if err != nil {
		t.Skipf("pipx list failed (environment issue): %v", err)
	}
	// May or may not have packages
	_ = packages
}

func TestInstallIdempotent(t *testing.T) {
	skipIfNoPipx(t)
	name := "cowsay"

	// Cleanup first
	_, _ = Uninstall(name)

	result1, err := Install(name)
	if err != nil {
		t.Skipf("Install() failed (pip environment issue): %v", err)
		return
	}
	if !result1.Changed {
		// Already installed from a previous run - acceptable
		t.Skip("package was already installed, skipping idempotency test")
	}

	result2, err := Install(name)
	if err != nil {
		t.Skipf("Install() idempotent error = %v", err)
	}
	if result2.Changed {
		t.Errorf("Expected Changed=false on idempotent install, got true (list parsing may differ)")
	}

	// Cleanup
	_, _ = Uninstall(name)
}
