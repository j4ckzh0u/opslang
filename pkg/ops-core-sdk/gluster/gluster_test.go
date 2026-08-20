package gluster

import (
	"testing"
)

func TestVolumeList(t *testing.T) {
	// May fail if gluster not installed
	_, _ = VolumeList()
}

func TestVolumeCreateEmptyName(t *testing.T) {
	_, err := VolumeCreate("", []string{"host:/brick"}, 0, 0, "")
	if err == nil {
		t.Fatal("VolumeCreate with empty name should return error")
	}
}

func TestVolumeCreateNoBricks(t *testing.T) {
	_, err := VolumeCreate("testvol", []string{}, 0, 0, "")
	if err == nil {
		t.Fatal("VolumeCreate with no bricks should return error")
	}
}

func TestVolumeDeleteEmptyName(t *testing.T) {
	_, err := VolumeDelete("")
	if err == nil {
		t.Fatal("VolumeDelete with empty name should return error")
	}
}

func TestVolumeStartEmptyName(t *testing.T) {
	_, err := VolumeStart("")
	if err == nil {
		t.Fatal("VolumeStart with empty name should return error")
	}
}

func TestVolumeStopEmptyName(t *testing.T) {
	_, err := VolumeStop("")
	if err == nil {
		t.Fatal("VolumeStop with empty name should return error")
	}
}

func TestPeerList(t *testing.T) {
	// May fail if gluster not installed
	_, _ = PeerList()
}

func TestPeerProbeEmptyHost(t *testing.T) {
	_, err := PeerProbe("")
	if err == nil {
		t.Fatal("PeerProbe with empty host should return error")
	}
}

func TestPeerDetachEmptyHost(t *testing.T) {
	_, err := PeerDetach("")
	if err == nil {
		t.Fatal("PeerDetach with empty host should return error")
	}
}
