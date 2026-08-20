// Package package_facts provides Ansible package_facts module equivalent.
package package_facts

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
)

// PackageFactsResult contains installed packages.
type PackageFactsResult struct {
	Packages map[string][]PackageInfo `json:"packages"`
	Error    string                   `json:"error,omitempty"`
}

// PackageInfo describes an installed package.
type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"` // "apt", "rpm", "pkgng"
	Arch    string `json:"arch,omitempty"`
}

// Collect gathers installed packages using available manager.
func Collect(managers []string) PackageFactsResult {
	if runtime.GOOS != "linux" {
		return PackageFactsResult{Error: "only supported on linux"}
	}
	if len(managers) == 0 {
		managers = []string{"auto"}
	}
	pkgs := map[string][]PackageInfo{}

	for _, m := range managers {
		switch m {
		case "auto":
			if apt := collectApt(); len(apt) > 0 {
				for k, v := range apt {
					pkgs[k] = v
				}
				return PackageFactsResult{Packages: pkgs}
			}
			if rpm := collectRpm(); len(rpm) > 0 {
				for k, v := range rpm {
					pkgs[k] = v
				}
				return PackageFactsResult{Packages: pkgs}
			}
		case "apt":
			pkgs = collectApt()
		case "rpm":
			pkgs = collectRpm()
		}
	}
	return PackageFactsResult{Packages: pkgs}
}

func collectApt() map[string][]PackageInfo {
	out, err := exec.Command("dpkg-query", "-W", "-f", "${Package}\t${Version}\t${Architecture}\n").Output()
	if err != nil {
		return nil
	}
	pkgs := map[string][]PackageInfo{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), "\t", 3)
		if len(parts) < 2 {
			continue
		}
		pi := PackageInfo{Name: parts[0], Version: parts[1], Source: "apt"}
		if len(parts) == 3 {
			pi.Arch = parts[2]
		}
		pkgs[pi.Name] = append(pkgs[pi.Name], pi)
	}
	return pkgs
}

func collectRpm() map[string][]PackageInfo {
	out, err := exec.Command("rpm", "-qa", "--queryformat", "%{NAME}\t%{VERSION}-%{RELEASE}\t%{ARCH}\n").Output()
	if err != nil {
		return nil
	}
	pkgs := map[string][]PackageInfo{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), "\t", 3)
		if len(parts) < 2 {
			continue
		}
		pi := PackageInfo{Name: parts[0], Version: parts[1], Source: "rpm"}
		if len(parts) == 3 {
			pi.Arch = parts[2]
		}
		pkgs[pi.Name] = append(pkgs[pi.Name], pi)
	}
	return pkgs
}

// JSON returns packages as JSON string.
func (r PackageFactsResult) JSON() string {
	b, _ := json.Marshal(r.Packages)
	return string(b)
}
