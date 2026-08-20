package podman

import (
	"os/exec"
	"testing"
)

func hasPodman() bool {
	_, err := exec.LookPath("podman")
	return err == nil
}

func TestRunEmptyImage(t *testing.T) {
	_, err := Run("", "", "")
	if err == nil {
		t.Fatal("expected error for empty image")
	}
}

func TestStopEmptyName(t *testing.T) {
	_, err := Stop("", 0)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStartEmptyName(t *testing.T) {
	_, err := Start("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemoveEmptyName(t *testing.T) {
	_, err := Remove("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInspectEmptyName(t *testing.T) {
	_, err := Inspect("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestPullEmptyImage(t *testing.T) {
	_, err := Pull("")
	if err == nil {
		t.Fatal("expected error for empty image")
	}
}

func TestRemoveImageEmptyID(t *testing.T) {
	_, err := RemoveImage("", false)
	if err == nil {
		t.Fatal("expected error for empty image ID")
	}
}

func TestCreatePodEmptyName(t *testing.T) {
	_, err := CreatePod("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStopPodEmptyName(t *testing.T) {
	_, err := StopPod("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemovePodEmptyName(t *testing.T) {
	_, err := RemovePod("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestListContainersNotAvailable(t *testing.T) {
	if !hasPodman() {
		_, err := ListContainers(false)
		if err == nil {
			t.Fatal("expected error when podman not found")
		}
	}
}

func TestListImagesNotAvailable(t *testing.T) {
	if !hasPodman() {
		_, err := ListImages()
		if err == nil {
			t.Fatal("expected error when podman not found")
		}
	}
}

func TestListPodsNotAvailable(t *testing.T) {
	if !hasPodman() {
		_, err := ListPods()
		if err == nil {
			t.Fatal("expected error when podman not found")
		}
	}
}
