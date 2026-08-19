// Kernel module management — Ansible modprobe equivalent.
package kernel

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ModuleInfo represents a loaded kernel module.
type ModuleInfo struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Used string `json:"used_by"`
}

// ModuleListResult is returned by ModuleList.
type ModuleListResult struct {
	Modules []ModuleInfo `json:"modules"`
	Count   int          `json:"count"`
}

// ModuleLoadResult is returned by ModuleLoad and ModuleUnload.
type ModuleLoadResult struct {
	Name    string `json:"name"`
	Loaded  bool   `json:"loaded"`
	Changed bool   `json:"changed"`
}

// ModuleList lists currently loaded kernel modules.
func ModuleList() (ModuleListResult, error) {
	result := ModuleListResult{Modules: make([]ModuleInfo, 0)}

	f, err := os.Open("/proc/modules")
	if err != nil {
		return result, fmt.Errorf("kernel.ModuleList: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			result.Modules = append(result.Modules, ModuleInfo{
				Name: fields[0],
				Size: fields[1],
				Used: fields[2],
			})
		}
	}
	result.Count = len(result.Modules)
	return result, nil
}

// ModuleLoad loads a kernel module using modprobe.
func ModuleLoad(name string) (ModuleLoadResult, error) {
	result := ModuleLoadResult{Name: name}

	// Check if already loaded
	loaded, err := isModuleLoaded(name)
	if err != nil {
		return result, err
	}
	if loaded {
		result.Loaded = true
		result.Changed = false
		return result, nil
	}

	cmd := exec.Command("modprobe", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return result, fmt.Errorf("kernel.ModuleLoad: modprobe %s failed: %s: %w", name, string(out), err)
	}
	result.Loaded = true
	result.Changed = true
	return result, nil
}

// ModuleUnload unloads a kernel module using modprobe -r.
func ModuleUnload(name string) (ModuleLoadResult, error) {
	result := ModuleLoadResult{Name: name}

	loaded, err := isModuleLoaded(name)
	if err != nil {
		return result, err
	}
	if !loaded {
		result.Loaded = false
		result.Changed = false
		return result, nil
	}

	cmd := exec.Command("modprobe", "-r", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return result, fmt.Errorf("kernel.ModuleUnload: modprobe -r %s failed: %s: %w", name, string(out), err)
	}
	result.Loaded = false
	result.Changed = true
	return result, nil
}

func isModuleLoaded(name string) (bool, error) {
	f, err := os.Open("/proc/modules")
	if err != nil {
		return false, fmt.Errorf("kernel: cannot read /proc/modules: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 1 && fields[0] == name {
			return true, nil
		}
	}
	return false, nil
}
