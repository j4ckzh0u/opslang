//go:build linux

package sys

import (
	"testing"
)

func TestListMounts(t *testing.T) {
	mounts, err := ListMounts()
	if err != nil {
		t.Fatalf("ListMounts failed: %v", err)
	}
	if len(mounts) == 0 {
		t.Fatal("expected at least one mount entry on Linux")
	}
	for _, m := range mounts {
		if m.Device == "" {
			t.Error("expected non-empty device")
		}
		if m.MountPoint == "" {
			t.Error("expected non-empty mount_point")
		}
		if m.FSType == "" {
			t.Error("expected non-empty fs_type")
		}
	}
}

func TestIsMounted(t *testing.T) {
	mounted, err := isMounted("/")
	if err != nil {
		t.Fatalf("isMounted failed: %v", err)
	}
	if !mounted {
		t.Fatal("expected / to be mounted")
	}

	mounted, err = isMounted("/nonexistent_mount_point_12345")
	if err != nil {
		t.Fatalf("isMounted failed: %v", err)
	}
	if mounted {
		t.Fatal("expected nonexistent path to not be mounted")
	}
}

func TestMount_EmptyDevice(t *testing.T) {
	_, err := Mount("", "/mnt/test", "", nil)
	if err == nil {
		t.Fatal("expected error for empty device")
	}
}

func TestMount_EmptyMountpoint(t *testing.T) {
	_, err := Mount("/dev/sda1", "", "", nil)
	if err == nil {
		t.Fatal("expected error for empty mountpoint")
	}
}

func TestUnmount_EmptyMountpoint(t *testing.T) {
	_, err := Unmount("")
	if err == nil {
		t.Fatal("expected error for empty mountpoint")
	}
}

func TestUnmount_NotMounted(t *testing.T) {
	res, err := Unmount("/nonexistent_mount_point_12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Changed {
		t.Fatal("expected changed=false for unmounted path")
	}
}

func TestListMounts_NonNil(t *testing.T) {
	mounts, err := ListMounts()
	if err != nil {
		t.Fatalf("ListMounts failed: %v", err)
	}
	if mounts == nil {
		t.Fatal("expected non-nil slice")
	}
}
