package sys

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
)

// pseudoFSTypes lists kernel pseudo-filesystems and volatile mounts that
// never hold operator-relevant data. Denylist wins over the device
// allowlist, so snap's "/dev/loop0 squashfs" is still excluded.
var pseudoFSTypes = map[string]bool{
	"proc": true, "sysfs": true, "devtmpfs": true, "devpts": true,
	"tmpfs": true, "ramfs": true, "overlay": true, "squashfs": true,
	"efivarfs": true, "bpf": true, "cgroup": true, "cgroup2": true,
	"cgroupfs": true, "autofs": true, "configfs": true, "debugfs": true,
	"tracefs": true, "pstore": true, "securityfs": true, "selinuxfs": true,
	"binfmt_misc": true, "mqueue": true, "hugetlbfs": true,
	"fusectl": true, "nsfs": true, "rpc_pipefs": true, "swap": true,
	"none": true,
}

// networkFSTypes lists filesystems whose device is not a local block
// device path but which still carry real data operators must monitor.
var networkFSTypes = map[string]bool{
	"nfs": true, "nfs4": true, "cifs": true, "smbfs": true,
	"ceph": true, "glusterfs": true, "afs": true, "9p": true,
}

// IsRealDataMount reports whether a mount carries operator-relevant data:
// local block devices (/dev/*), network storage, or ZFS datasets. Pseudo
// filesystems are excluded even when their device looks like /dev/*
// (snap loop mounts). Pure function so the rules stay unit-testable
// without a real machine's mount table.
func IsRealDataMount(device, fstype string) bool {
	fstype = strings.ToLower(strings.TrimSpace(fstype))
	if pseudoFSTypes[fstype] {
		return false
	}
	if networkFSTypes[fstype] || fstype == "zfs" {
		return true
	}
	return strings.HasPrefix(device, "/dev/")
}

// GetDiskPartitions returns only real data-bearing mounts: physical or
// virtual block devices plus network storage, with per-mount capacity.
// Kernel pseudo-filesystems (tmpfs, overlay, squashfs, proc, ...) and
// snap loop mounts are filtered out on purpose: servers differ wildly in
// how many of those they mount, and scripts should not need to know.
//
// Capacity fields stay zero when a single mount cannot be stat'ed (for
// example a stale NFS export) - that one entry degrades, the call as a
// whole does not fail.
//
// For the unfiltered raw mount table use ListMounts().
func GetDiskPartitions() ([]DiskPartition, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}
	result := make([]DiskPartition, 0, len(parts))
	for _, p := range parts {
		if !IsRealDataMount(p.Device, p.Fstype) {
			continue
		}
		entry := DiskPartition{
			Device:     p.Device,
			Mountpoint: p.Mountpoint,
			Fstype:     p.Fstype,
			Opts:       strings.Join(p.Opts, ","),
		}
		if usage, err := disk.Usage(p.Mountpoint); err == nil {
			entry.TotalBytes = usage.Total
			entry.UsedBytes = usage.Used
			entry.FreeBytes = usage.Free
			entry.UsedPercent = usage.UsedPercent
		}
		result = append(result, entry)
	}
	return result, nil
}

// VirtInfo answers "what kind of machine am I on": container, VM, or
// bare metal. Placement decisions (agent install, backup strategy,
// resource sizing) all start from this question.
type VirtInfo struct {
	// System is the hypervisor/container runtime identifier as reported
	// by the OS (e.g. "docker", "kvm", "vmware", "xen"); empty means the
	// probe found no virtualization layer (bare metal).
	System string `json:"system"`
	// Role distinguishes "guest" (virtualized) from "host".
	Role string `json:"role"`
	// IsContainer collapses the container-runtime list into the one bit
	// scripts actually branch on.
	IsContainer bool `json:"is_container"`
}

var containerSystems = map[string]bool{
	"docker": true, "podman": true, "container": true, "lxc": true,
	"lxc-libvirt": true, "systemd-nspawn": true, "rkt": true,
}

// GetVirtInfo classifies the execution environment. The raw probe can
// legitimately fail on exotic platforms without making the machine
// class unknown-in-a-useful-way, so failures surface as an explicit
// error rather than a guessed default.
func GetVirtInfo() (VirtInfo, error) {
	system, role, err := host.Virtualization()
	if err != nil {
		return VirtInfo{}, fmt.Errorf("failed to detect virtualization: %w", err)
	}
	return VirtInfo{
		System:      system,
		Role:        role,
		IsContainer: containerSystems[strings.ToLower(system)],
	}, nil
}
