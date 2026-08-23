package opspkg

import (
	"fmt"
	"strings"
)

// OwnerResult is returned by Owner: which installed package (if any)
// claims a given file path.
type OwnerResult struct {
	// File is the queried path, echoed back verbatim.
	File string `json:"file"`
	// Package is the owning package name; empty when not owned by any
	// package (self-compiled binaries, snap-mounted files, files placed
	// by hand — a real and important ops distinction, not an error).
	Package string `json:"package"`
	Found   bool   `json:"found"`
	Manager string `json:"manager"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Owner answers "which package installed this file" — the ops question
// behind "this process's binary, who put it there?". It maps to
// `dpkg -S <path>` (apt) or `rpm -qf <path>` (yum/dnf), with structured
// parsing of the field-delimited output. A file owned by no package is a
// normal finding (found=false, no error); only execution failures error.
func Owner(path string) (OwnerResult, error) {
	if strings.TrimSpace(path) == "" {
		return OwnerResult{File: path, Error: "file path is required"}, fmt.Errorf("file path is required")
	}

	mgrName, mgrPath, err := detectManager()
	if err != nil {
		return OwnerResult{File: path, Error: err.Error()}, err
	}

	var owner string
	var output string
	var found bool
	switch mgrName {
	case "apt":
		// dpkg -S matches the path as recorded in the database. On
		// usrmerge systems /usr/bin/ls is recorded as /bin/ls, so a
		// miss falls back to the legacy-path variants — the same trap
		// every dpkg -S user hits once.
		candidates := usrmergeVariants(path)
		for _, cand := range candidates {
			out, execErr := runCommand("dpkg", "-S", cand)
			if execErr != nil {
				continue // "no path found" — try the next variant
			}
			if pkgName := parseDpkgSearchOwner(out); pkgName != "" {
				owner, output, found = pkgName, out, true
				break
			}
		}
	case "yum", "dnf":
		out, execErr := runCommand(mgrPath, "-qf", "--queryformat", "%{NAME}", path)
		output = out
		if execErr == nil {
			owner = strings.TrimSpace(out)
			found = owner != ""
		}
	default:
		err := fmt.Errorf("unsupported manager: %s", mgrName)
		return OwnerResult{File: path, Manager: mgrName, Error: err.Error()}, err
	}

	res := OwnerResult{File: path, Manager: mgrName, Output: strings.TrimSpace(output)}
	if found && owner != "" {
		res.Package = owner
		res.Found = true
	}
	return res, nil
}

// usrmergeVariants returns path followed by its pre-usrmerge legacy forms:
// /usr/bin/x -> /bin/x, /usr/sbin/x -> /sbin/x, /usr/lib/x -> /lib/x.
// On non-usrmerge systems the legacy form simply misses and the original
// path has already been tried first.
func usrmergeVariants(path string) []string {
	out := []string{path}
	for _, prefix := range []string{"/usr/bin/", "/usr/sbin/", "/usr/lib/"} {
		if strings.HasPrefix(path, prefix) {
			out = append(out, "/"+path[len("/usr/"):])
			break
		}
	}
	return out
}

// parseDpkgSearchOwner extracts the first owning package from dpkg -S
// output. Real-world shapes handled:
//
//	systemd: /lib/systemd/systemd
//	libc6-dev, linux-libc-dev: /usr/include/x86_64-linux-gnu/gnu
//	diversion by iproute2 from: /sbin/route\niproute2: /sbin/route.real
//	dpkg-query: no path found matching pattern /tmp/selfbuilt
func parseDpkgSearchOwner(output string) string {
	var owner string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "diversion") ||
			strings.HasPrefix(line, "locale:") ||
			strings.HasPrefix(line, "dpkg-query:") || strings.HasPrefix(line, "dpkg:") {
			continue
		}
		idx := strings.LastIndex(line, ": ")
		if idx <= 0 {
			continue
		}
		names := line[:idx]
		first := names
		if c := strings.Index(names, ","); c >= 0 {
			first = names[:c]
		}
		if strings.ContainsAny(first, " \t") {
			// Not a package-name field.
			continue
		}
		owner = strings.TrimSpace(first)
	}
	return owner
}
