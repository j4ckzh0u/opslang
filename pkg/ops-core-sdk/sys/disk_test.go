package sys

import (
	"strings"
	"testing"
)

// TestIsRealDataMount pins the semantic filter rules against a synthetic
// mount table: no real machine needed, and every rule (denylist wins,
// network fs allowed, ZFS device naming) has an explicit case.
func TestIsRealDataMount(t *testing.T) {
	cases := []struct {
		name   string
		device string
		fstype string
		want   bool
	}{
		{"sata disk", "/dev/sda1", "ext4", true},
		{"nvme disk", "/dev/nvme0n1p2", "xfs", true},
		{"virtio disk", "/dev/vda1", "ext4", true},
		{"lvm volume", "/dev/mapper/vg-root", "ext4", true},
		{"md raid", "/dev/md0", "ext4", true},
		{"mmc storage", "/dev/mmcblk0p1", "vfat", true},
		{"nfs export", "server:/export/data", "nfs4", true},
		{"cifs share", "//fileserver/share", "cifs", true},
		{"zfs dataset (no /dev prefix)", "rpool/ROOT/ubuntu", "zfs", true},
		{"tmpfs on /dev-style path is still pseudo", "/dev/shm", "tmpfs", false},
		{"snap loop squashfs", "/dev/loop0", "squashfs", false},
		{"proc", "proc", "proc", false},
		{"sysfs", "sysfs", "sysfs", false},
		{"devtmpfs", "udev", "devtmpfs", false},
		{"overlay container root", "overlay", "overlay", false},
		{"efivarfs", "efivarfs", "efivarfs", false},
		{"cgroup2", "cgroup2", "cgroup2", false},
		{"autofs", "systemd-1", "autofs", false},
		{"fstype case-insensitive", "/dev/sda1", "EXT4", true},
		{"pseudo fstype case-insensitive", "/dev/pstore", "TMPFS", false},
		{"empty everything", "", "", false},
		{"unknown fstype without /dev device", "weird-thing", "fancyfs", false},
	}
	for _, tc := range cases {
		if got := IsRealDataMount(tc.device, tc.fstype); got != tc.want {
			t.Errorf("%s: IsRealDataMount(%q, %q) = %v, want %v",
				tc.name, tc.device, tc.fstype, got, tc.want)
		}
	}
}

// TestGetDiskPartitionsFiltered runs the real call on the current machine
// and asserts the invariants that must hold wherever it executes. On a
// container host with only overlay mounts this legitimately returns an
// empty list - callers wanting everything have ListMounts().
func TestGetDiskPartitionsFiltered(t *testing.T) {
	parts, err := GetDiskPartitions()
	if err != nil {
		t.Fatalf("GetDiskPartitions() error = %v", err)
	}
	banned := map[string]bool{
		"tmpfs": true, "proc": true, "sysfs": true, "devtmpfs": true,
		"overlay": true, "squashfs": true, "efivarfs": true,
		"cgroup": true, "cgroup2": true, "devpts": true, "mqueue": true,
		"hugetlbfs": true,
	}
	sawRoot := false
	for _, p := range parts {
		if banned[strings.ToLower(p.Fstype)] {
			t.Errorf("pseudo filesystem leaked into result: %+v", p)
		}
		if !strings.HasPrefix(p.Device, "/dev/") && !IsRealDataMount(p.Device, p.Fstype) {
			t.Errorf("entry fails its own inclusion rule: %+v", p)
		}
		if p.Mountpoint == "" {
			t.Errorf("empty mountpoint in result: %+v", p)
		}
		if p.Mountpoint == "/" {
			sawRoot = true
			if p.TotalBytes == 0 {
				t.Errorf("root mount has no capacity info: %+v", p)
			}
			if p.FreeBytes >= p.TotalBytes && p.TotalBytes > 0 {
				t.Errorf("free (%d) larger than total (%d)", p.FreeBytes, p.TotalBytes)
			}
		}
	}
	if !sawRoot {
		t.Skip("current machine exposes no / mount through the filter (container?)")
	}
}

// TestGetVirtInfo runs the real probe and asserts the classification
// invariants: role is one of the known values, and IsContainer agrees
// with System for every container runtime.
func TestGetVirtInfo(t *testing.T) {
	info, err := GetVirtInfo()
	if err != nil {
		// Some platforms (BSD jails, exotic CI) cannot probe; an explicit
		// error is the honest contract - just verify it is populated.
		if info.System != "" {
			t.Errorf("error return must not carry partial data: %+v", info)
		}
		t.Skipf("virtualization probe unsupported here: %v", err)
	}
	switch info.Role {
	case "guest", "host", "":
		// known values; empty only when no virtualization layer exists
	default:
		t.Errorf("unexpected role %q", info.Role)
	}
	if info.IsContainer != containerSystems[strings.ToLower(info.System)] {
		t.Errorf("IsContainer inconsistent with system %q", info.System)
	}
}
