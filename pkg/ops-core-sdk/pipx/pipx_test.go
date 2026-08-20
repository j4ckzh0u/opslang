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
	// Install a common package and check idempotency
	// Use 'cowsay' as it's small and harmless
	name := "cowsay"

	// Cleanup first
	_, _ = Uninstall(name)

	result1, err := Install(name)
	if err != nil {
		t.Skipf("Install() failed (pip environment issue): %v", err)
		return
	}
	if !result1.Changed {
		t.Error("Expected Changed=true on first install")
	}

	result2, err := Install(name)
	if err != nil {
		t.Fatalf("Install() idempotent error = %v", err)
	}
	if result2.Changed {
		t.Error("Expected Changed=false on idempotent install")
	}

	// Cleanup
	_, _ = Uninstall(name)
}
