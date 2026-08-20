package btrfs

import (
	"testing"
)

func TestSubvolumeCreateEmptyPath(t *testing.T) {
	_, err := SubvolumeCreate("")
	if err == nil {
		t.Fatal("SubvolumeCreate('') should return error for empty path")
	}
}

func TestSubvolumeDeleteEmptyPath(t *testing.T) {
	_, err := SubvolumeDelete("")
	if err == nil {
		t.Fatal("SubvolumeDelete('') should return error for empty path")
	}
}

func TestSnapshotCreateEmptyPaths(t *testing.T) {
	_, err := SnapshotCreate("", "", false)
	if err == nil {
		t.Fatal("SnapshotCreate('', '') should return error for empty paths")
	}
}

func TestDeviceAddEmptyPaths(t *testing.T) {
	_, err := DeviceAdd("", "")
	if err == nil {
		t.Fatal("DeviceAdd('', '') should return error for empty paths")
	}
}

func TestDeviceRemoveEmptyPaths(t *testing.T) {
	_, err := DeviceRemove("", "")
	if err == nil {
		t.Fatal("DeviceRemove('', '') should return error for empty paths")
	}
}
