package file

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// EnsureResult is returned by Ensure. Actions lists the filesystem mutations
// that actually executed, so an idempotent re-run reports an empty list with
// changed=false.
type EnsureResult struct {
	Path    string   `json:"path"`
	State   string   `json:"state"`
	Type    string   `json:"type"`
	Mode    string   `json:"mode,omitempty"`
	Changed bool     `json:"changed"`
	Actions []string `json:"actions"`
	Message string   `json:"message"`
	Error   string   `json:"error,omitempty"`
}

// Ensure converges a path to the requested state, mirroring the Ansible
// file module:
//
//	state=directory  missing -> mkdir -p (mode defaults 0755)
//	                  exists as dir -> maybe chmod
//	                  exists as file -> error (never silently replace)
//	state=file       exists as regular file -> maybe chmod
//	                  missing -> error (state=file does not create, like Ansible)
//	state=touch      missing -> create empty file; existing -> left untouched
//	                  (deliberate divergence from Ansible, which refreshes
//	                  mtime on every run: convergence runs stay idempotent)
//	state=absent     exists -> removed (directories recursively); else no-op
//
// mode is an octal string like "0755" and is enforced on the 9 permission
// bits; an empty mode leaves permissions as they are.
func Ensure(path, state, mode string) (EnsureResult, error) {
	state = strings.ToLower(strings.TrimSpace(state))
	res := EnsureResult{Path: path, State: state, Actions: []string{}}

	wantMode, err := parseOctalMode(mode)
	if err != nil {
		res.Error = err.Error()
		return res, fmt.Errorf("file.ensure: %w", err)
	}

	info, statErr := os.Lstat(path)
	exists := statErr == nil
	switch {
	case statErr == nil:
		if info.IsDir() {
			res.Type = "directory"
		} else if info.Mode().IsRegular() {
			res.Type = "file"
		} else {
			res.Type = "other"
		}
	case os.IsNotExist(statErr):
		res.Type = ""
	default:
		res.Error = statErr.Error()
		res.Message = "unable to stat path"
		return res, fmt.Errorf("file.ensure stat %s: %w", path, statErr)
	}

	switch state {
	case "directory":
		return ensureDirectory(res, path, exists, wantMode)
	case "file":
		return ensureRegularFile(res, path, exists, wantMode)
	case "touch":
		return ensureTouch(res, path, exists, wantMode)
	case "absent":
		return ensureAbsent(res, path, exists)
	case "":
		res.Error = "state is required: directory|file|touch|absent"
		return res, fmt.Errorf("file.ensure: %s", res.Error)
	default:
		res.Error = fmt.Sprintf("invalid state %q: must be directory|file|touch|absent", state)
		return res, fmt.Errorf("file.ensure: %s", res.Error)
	}
}

func ensureDirectory(res EnsureResult, path string, exists bool, wantMode uint32) (EnsureResult, error) {
	if exists && res.Type == "file" {
		res.Error = "path exists and is a regular file; refusing to replace it with a directory"
		res.Message = "type conflict"
		return res, fmt.Errorf("file.ensure: %s", res.Error)
	}

	if !exists {
		mkdirMode := os.FileMode(0755)
		if wantMode != 0 {
			mkdirMode = os.FileMode(wantMode)
		}
		if err := os.MkdirAll(path, mkdirMode); err != nil {
			res.Error = err.Error()
			res.Message = "mkdir failed"
			return res, fmt.Errorf("file.ensure mkdir %s: %w", path, err)
		}
		res.Type = "directory"
		res.Changed = true
		res.Actions = append(res.Actions, "mkdir")
		res.Message = "directory created"
		applyModeEcho(&res, wantMode)
		return res, nil
	}

	// Directory already present: only mode can still drift.
	if changed, err := convergeMode(&res, path, wantMode); err != nil {
		return res, err
	} else if changed {
		res.Message = "directory mode converged"
		return res, nil
	}
	res.Message = "directory already up to date"
	return res, nil
}

func ensureRegularFile(res EnsureResult, path string, exists bool, wantMode uint32) (EnsureResult, error) {
	if !exists {
		res.Error = "path does not exist; state=file never creates (use state=touch)"
		res.Message = "missing file"
		return res, fmt.Errorf("file.ensure: %s", res.Error)
	}
	if res.Type != "file" {
		res.Error = fmt.Sprintf("path is a %s, not a regular file", res.Type)
		res.Message = "type conflict"
		return res, fmt.Errorf("file.ensure: %s", res.Error)
	}

	if changed, err := convergeMode(&res, path, wantMode); err != nil {
		return res, err
	} else if changed {
		res.Message = "file mode converged"
		return res, nil
	}
	res.Message = "file already up to date"
	return res, nil
}

func ensureTouch(res EnsureResult, path string, exists bool, wantMode uint32) (EnsureResult, error) {
	if exists {
		if res.Type != "file" {
			res.Error = fmt.Sprintf("path is a %s, not a regular file", res.Type)
			res.Message = "type conflict"
			return res, fmt.Errorf("file.ensure: %s", res.Error)
		}
		if changed, err := convergeMode(&res, path, wantMode); err != nil {
			return res, err
		} else if changed {
			res.Message = "file mode converged"
			return res, nil
		}
		res.Message = "file already present"
		return res, nil
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		res.Error = err.Error()
		res.Message = "create failed"
		return res, fmt.Errorf("file.ensure touch %s: %w", path, err)
	}
	f.Close()
	res.Type = "file"
	res.Changed = true
	res.Actions = append(res.Actions, "create")
	res.Message = "file created"
	if wantMode != 0 {
		if err := os.Chmod(path, os.FileMode(wantMode)); err != nil {
			res.Error = err.Error()
			return res, fmt.Errorf("file.ensure chmod %s: %w", path, err)
		}
		res.Actions = append(res.Actions, "chmod")
	}
	applyModeEcho(&res, wantMode)
	return res, nil
}

func ensureAbsent(res EnsureResult, path string, exists bool) (EnsureResult, error) {
	if !exists {
		res.Type = ""
		res.Message = "path already absent"
		return res, nil
	}

	if err := os.RemoveAll(path); err != nil {
		res.Error = err.Error()
		res.Message = "remove failed"
		return res, fmt.Errorf("file.ensure remove %s: %w", path, err)
	}
	res.Changed = true
	res.Actions = append(res.Actions, "remove")
	res.Type = ""
	res.Message = "path removed"
	return res, nil
}

// convergeMode chmods the path when the 9 permission bits differ from
// wantMode. A zero wantMode means "leave permissions alone".
func convergeMode(res *EnsureResult, path string, wantMode uint32) (bool, error) {
	if wantMode == 0 {
		return false, nil
	}
	info, err := os.Lstat(path)
	if err != nil {
		res.Error = err.Error()
		return false, fmt.Errorf("file.ensure stat %s: %w", path, err)
	}
	if info.Mode().Perm() == os.FileMode(wantMode).Perm() {
		res.Mode = fmt.Sprintf("%04o", info.Mode().Perm())
		return false, nil
	}
	if err := os.Chmod(path, os.FileMode(wantMode)); err != nil {
		res.Error = err.Error()
		return false, fmt.Errorf("file.ensure chmod %s: %w", path, err)
	}
	res.Mode = fmt.Sprintf("%04o", wantMode)
	res.Changed = true
	res.Actions = append(res.Actions, "chmod")
	return true, nil
}

func applyModeEcho(res *EnsureResult, wantMode uint32) {
	if wantMode != 0 {
		res.Mode = fmt.Sprintf("%04o", wantMode)
	}
}

// parseOctalMode parses "0755"-style permission strings. Empty is valid and
// means "do not touch permissions".
func parseOctalMode(mode string) (uint32, error) {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return 0, nil
	}
	if len(mode) > 4 {
		return 0, fmt.Errorf("invalid mode %q: expected up to 4 octal digits", mode)
	}
	v, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid mode %q: %w", mode, err)
	}
	if v > 0777 {
		return 0, fmt.Errorf("invalid mode %q: only permission bits (0-0777) are supported", mode)
	}
	return uint32(v), nil
}
