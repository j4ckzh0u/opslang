// Package zfs provides ZFS filesystem management operations.
package zfs

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a ZFS operation.
type ActionResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DatasetInfo represents information about a ZFS dataset.
type DatasetInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Used       string `json:"used"`
	Avail      string `json:"avail"`
	Refer      string `json:"refer"`
	Mountpoint string `json:"mountpoint"`
}

// ListResult represents the result of listing ZFS datasets.
type ListResult struct {
	Datasets []DatasetInfo `json:"datasets"`
}

// PoolInfo represents information about a ZFS pool.
type PoolInfo struct {
	Name      string `json:"name"`
	Size      string `json:"size"`
	Used      string `json:"used"`
	Avail     string `json:"avail"`
	Health    string `json:"health"`
	AltRoot   string `json:"alt_root"`
}

// PoolListResult represents the result of listing ZFS pools.
type PoolListResult struct {
	Pools []PoolInfo `json:"pools"`
}

// Create creates a new ZFS dataset.
func Create(name string, properties map[string]string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("dataset name is required")
	}

	args := []string{"create"}
	for k, v := range properties {
		args = append(args, "-o", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, name)

	out, err := exec.Command("zfs", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("zfs create failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Destroy destroys a ZFS dataset.
func Destroy(name string, recursive bool) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("dataset name is required")
	}

	args := []string{"destroy"}
	if recursive {
		args = append(args, "-r")
	}
	args = append(args, name)

	out, err := exec.Command("zfs", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("zfs destroy failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Set sets a property on a ZFS dataset.
func Set(name string, property string, value string) (ActionResult, error) {
	if name == "" || property == "" {
		return ActionResult{}, fmt.Errorf("dataset name and property are required")
	}

	args := []string{"set", fmt.Sprintf("%s=%s", property, value), name}
	out, err := exec.Command("zfs", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("zfs set failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Get gets a property from a ZFS dataset.
func Get(name string, property string) (map[string]string, error) {
	if name == "" {
		return nil, fmt.Errorf("dataset name is required")
	}

	args := []string{"get", "-H", "-p"}
	if property != "" {
		args = append(args, property)
	} else {
		args = append(args, "all")
	}
	args = append(args, name)

	out, err := exec.Command("zfs", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("zfs get failed: %w (output: %s)", err, string(out))
	}

	result := make(map[string]string)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Format: name\tproperty\tvalue\tsource
		fields := strings.Split(line, "\t")
		if len(fields) >= 3 {
			result[fields[1]] = fields[2]
		}
	}
	return result, nil
}

// List lists ZFS datasets.
func List() (ListResult, error) {
	out, err := exec.Command("zfs", "list", "-H", "-p", "-o", "name,type,used,avail,refer,mountpoint").CombinedOutput()
	if err != nil {
		return ListResult{}, fmt.Errorf("zfs list failed: %w (output: %s)", err, string(out))
	}

	result := ListResult{Datasets: make([]DatasetInfo, 0)}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) >= 6 {
			ds := DatasetInfo{
				Name:       fields[0],
				Type:       fields[1],
				Used:       fields[2],
				Avail:      fields[3],
				Refer:      fields[4],
				Mountpoint: fields[5],
			}
			result.Datasets = append(result.Datasets, ds)
		}
	}
	return result, nil
}

// Exists checks if a ZFS dataset exists.
func Exists(name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("dataset name is required")
	}

	err := exec.Command("zfs", "list", name).Run()
	return err == nil, nil
}

// ListPools lists ZFS storage pools.
func ListPools() (PoolListResult, error) {
	out, err := exec.Command("zpool", "list", "-H", "-p", "-o", "name,size,used,avail,health,altroot").CombinedOutput()
	if err != nil {
		return PoolListResult{}, fmt.Errorf("zpool list failed: %w (output: %s)", err, string(out))
	}

	result := PoolListResult{Pools: make([]PoolInfo, 0)}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) >= 5 {
			pool := PoolInfo{
				Name:   fields[0],
				Size:   fields[1],
				Used:   fields[2],
				Avail:  fields[3],
				Health: fields[4],
			}
			if len(fields) >= 6 {
				pool.AltRoot = fields[5]
			}
			result.Pools = append(result.Pools, pool)
		}
	}
	return result, nil
}

// GetPoolStatus gets the status of a ZFS pool.
func GetPoolStatus(name string) (map[string]interface{}, error) {
	args := []string{"status", "-j"}
	if name != "" {
		args = append(args, name)
	}

	out, err := exec.Command("zpool", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("zpool status failed: %w (output: %s)", err, string(out))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		// Fallback: return raw output
		return map[string]interface{}{"raw": string(out)}, nil
	}
	return result, nil
}

// Snapshot creates a snapshot of a ZFS dataset.
func Snapshot(name string, snapshotName string) (ActionResult, error) {
	if name == "" || snapshotName == "" {
		return ActionResult{}, fmt.Errorf("dataset name and snapshot name are required")
	}

	fullName := fmt.Sprintf("%s@%s", name, snapshotName)
	out, err := exec.Command("zfs", "snapshot", fullName).CombinedOutput()
	if err != nil {
		return ActionResult{Name: fullName, Success: false, Error: string(out)}, fmt.Errorf("zfs snapshot failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: fullName, Changed: true, Success: true}, nil
}

// DestroySnapshot destroys a ZFS snapshot.
func DestroySnapshot(name string, snapshotName string) (ActionResult, error) {
	if name == "" || snapshotName == "" {
		return ActionResult{}, fmt.Errorf("dataset name and snapshot name are required")
	}

	fullName := fmt.Sprintf("%s@%s", name, snapshotName)
	out, err := exec.Command("zfs", "destroy", fullName).CombinedOutput()
	if err != nil {
		return ActionResult{Name: fullName, Success: false, Error: string(out)}, fmt.Errorf("zfs destroy snapshot failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: fullName, Changed: true, Success: true}, nil
}
