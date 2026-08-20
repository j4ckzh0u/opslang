package docker_volume

import (
	"os/exec"
	"testing"
)

func skipIfNoDocker(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found in PATH")
	}
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("docker daemon not accessible")
	}
}

func TestCreateValidation(t *testing.T) {
	_, err := Create("", "local")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRemoveValidation(t *testing.T) {
	_, err := Remove("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestInspectValidation(t *testing.T) {
	_, err := Inspect("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestCreateAndRemove(t *testing.T) {
	skipIfNoDocker(t)
	name := "opslang-test-vol"

	// Cleanup first
	_, _ = Remove(name)

	// Create
	result, err := Create(name, "local")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !result.Changed {
		t.Error("Expected Changed=true on create")
	}

	// Create again - should be idempotent
	result2, err := Create(name, "local")
	if err != nil {
		t.Fatalf("Create() idempotent error = %v", err)
	}
	if result2.Changed {
		t.Error("Expected Changed=false on idempotent create")
	}

	// Inspect
	inspResult, err := Inspect(name)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if inspResult.Volume == nil || inspResult.Volume.Name != name {
		t.Error("Inspect() returned wrong volume info")
	}

	// Remove
	rmResult, err := Remove(name)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if !rmResult.Changed {
		t.Error("Expected Changed=true on remove")
	}

	// Remove again - should be idempotent
	rmResult2, err := Remove(name)
	if err != nil {
		t.Fatalf("Remove() idempotent error = %v", err)
	}
	if rmResult2.Changed {
		t.Error("Expected Changed=false on idempotent remove")
	}
}

func TestList(t *testing.T) {
	skipIfNoDocker(t)
	volumes, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	// May or may not have volumes, just check no error
	_ = volumes
}
