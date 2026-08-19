// Package authorized_key manages SSH authorized_keys files.
// Equivalent to ansible.posix.authorized_key module.
package authorized_key

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ManageResult is returned by Manage.
type ManageResult struct {
	Status  string `json:"status"`            // success/failed
	Changed bool   `json:"changed"`           // whether the key was added/removed
	Key     string `json:"key,omitempty"`     // the key fingerprint (truncated)
	User    string `json:"user"`              // target user
	Path    string `json:"path,omitempty"`    // authorized_keys path
	Keys    int    `json:"keys"`              // total keys after operation
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Status string   `json:"status"`
	Keys   []string `json:"keys"` // key entries (comments or truncated)
	User   string   `json:"user"`
	Count  int      `json:"count"`
	Error  string   `json:"error,omitempty"`
}

// CheckResult is returned by Check.
type CheckResult struct {
	Status   string `json:"status"`
	Found    bool   `json:"found"`
	User     string `json:"user"`
	KeyMatch string `json:"key_match,omitempty"`
	Error    string `json:"error,omitempty"`
}

func sshDir(u *user.User) string {
	return filepath.Join(u.HomeDir, ".ssh")
}

func authKeysPath(u *user.User, path string) string {
	if path != "" {
		return path
	}
	return filepath.Join(sshDir(u), "authorized_keys")
}

func resolveUser(username string) (*user.User, error) {
	if username == "" {
		return user.Current()
	}
	return user.Lookup(username)
}

func readKeys(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var keys []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		keys = append(keys, line)
	}
	return keys, scanner.Err()
}

func writeKeys(path string, keys []string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create ssh dir: %w", err)
	}

	content := strings.Join(keys, "\n")
	if len(keys) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0600)
}

// Manage adds or removes an SSH authorized key for a user.
// state: "present" or "absent".
func Manage(username, key, state, path string) ManageResult {
	u, err := resolveUser(username)
	if err != nil {
		return ManageResult{Status: "failed", Error: fmt.Sprintf("resolve user: %v", err)}
	}

	if key == "" {
		return ManageResult{Status: "failed", User: u.Username, Error: "key is required"}
	}

	keysPath := authKeysPath(u, path)
	keys, err := readKeys(keysPath)
	if err != nil {
		return ManageResult{Status: "failed", User: u.Username, Error: fmt.Sprintf("read keys: %v", err)}
	}

	// Normalize: extract the key type + base64 part for comparison
	keyFields := strings.Fields(key)
	keyID := ""
	if len(keyFields) >= 2 {
		keyID = keyFields[0] + " " + keyFields[1]
	} else {
		keyID = key
	}

	found := false
	for _, k := range keys {
		kf := strings.Fields(k)
		kID := k
		if len(kf) >= 2 {
			kID = kf[0] + " " + kf[1]
		}
		if kID == keyID {
			found = true
			break
		}
	}

	fingerprint := keyID
	if len(fingerprint) > 40 {
		fingerprint = fingerprint[:40] + "..."
	}

	switch state {
	case "present":
		if found {
			return ManageResult{
				Status: "success", Changed: false,
				Key: fingerprint, User: u.Username, Path: keysPath, Keys: len(keys),
			}
		}
		keys = append(keys, key)
		if err := writeKeys(keysPath, keys); err != nil {
			return ManageResult{Status: "failed", User: u.Username, Error: fmt.Sprintf("write keys: %v", err)}
		}
		return ManageResult{
			Status: "success", Changed: true,
			Key: fingerprint, User: u.Username, Path: keysPath, Keys: len(keys),
		}

	case "absent":
		if !found {
			return ManageResult{
				Status: "success", Changed: false,
				Key: fingerprint, User: u.Username, Path: keysPath, Keys: len(keys),
			}
		}
		var filtered []string
		for _, k := range keys {
			kf := strings.Fields(k)
			kID := k
			if len(kf) >= 2 {
				kID = kf[0] + " " + kf[1]
			}
			if kID != keyID {
				filtered = append(filtered, k)
			}
		}
		if err := writeKeys(keysPath, filtered); err != nil {
			return ManageResult{Status: "failed", User: u.Username, Error: fmt.Sprintf("write keys: %v", err)}
		}
		return ManageResult{
			Status: "success", Changed: true,
			Key: fingerprint, User: u.Username, Path: keysPath, Keys: len(filtered),
		}

	default:
		return ManageResult{Status: "failed", User: u.Username, Error: "state must be 'present' or 'absent'"}
	}
}

// List lists all SSH authorized keys for a user.
func List(username, path string) ListResult {
	u, err := resolveUser(username)
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("resolve user: %v", err)}
	}

	keysPath := authKeysPath(u, path)
	keys, err := readKeys(keysPath)
	if err != nil {
		return ListResult{Status: "failed", User: u.Username, Error: fmt.Sprintf("read keys: %v", err)}
	}

	return ListResult{
		Status: "success",
		Keys:   keys,
		User:   u.Username,
		Count:  len(keys),
	}
}

// Check checks if a specific key exists in the authorized_keys file.
func Check(username, key, path string) CheckResult {
	u, err := resolveUser(username)
	if err != nil {
		return CheckResult{Status: "failed", Error: fmt.Sprintf("resolve user: %v", err)}
	}

	if key == "" {
		return CheckResult{Status: "failed", User: u.Username, Error: "key is required"}
	}

	keysPath := authKeysPath(u, path)
	keys, err := readKeys(keysPath)
	if err != nil {
		return CheckResult{Status: "failed", User: u.Username, Error: fmt.Sprintf("read keys: %v", err)}
	}

	keyFields := strings.Fields(key)
	keyID := key
	if len(keyFields) >= 2 {
		keyID = keyFields[0] + " " + keyFields[1]
	}

	for _, k := range keys {
		kf := strings.Fields(k)
		kID := k
		if len(kf) >= 2 {
			kID = kf[0] + " " + kf[1]
		}
		if kID == keyID {
			return CheckResult{Status: "success", Found: true, User: u.Username, KeyMatch: keyID}
		}
	}

	return CheckResult{Status: "success", Found: false, User: u.Username}
}
