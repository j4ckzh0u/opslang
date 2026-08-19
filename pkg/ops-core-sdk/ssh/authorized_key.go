// SSH authorized_key management — Ansible authorized_key module equivalent.
package ssh

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AuthorizedKeyResult is returned by AuthorizedKey.
type AuthorizedKeyResult struct {
	User    string `json:"user"`
	Key     string `json:"key"`
	State   string `json:"state"`
	Changed bool   `json:"changed"`
}

// AuthorizedKeyAdd adds a public key to a user's authorized_keys file.
func AuthorizedKeyAdd(user string, key string, exclusive bool) (AuthorizedKeyResult, error) {
	result := AuthorizedKeyResult{User: user, Key: key, State: "present"}

	homeDir, err := userHome(user)
	if err != nil {
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	authFile := filepath.Join(sshDir, "authorized_keys")

	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}

	data, err := os.ReadFile(authFile)
	if err != nil && !os.IsNotExist(err) {
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}

	keyParts := strings.Fields(key)
	if len(keyParts) < 2 {
		return result, fmt.Errorf("ssh.AuthorizedKey: invalid key format")
	}
	keyType := keyParts[0]
	keyData := keyParts[1]

	lines := strings.Split(string(data), "\n")
	found := false
	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 && fields[0] == keyType && fields[1] == keyData {
			found = true
			if exclusive {
				continue // skip this key (remove it)
			}
			newLines = append(newLines, line)
		} else {
			if exclusive {
				continue // remove all other keys
			}
			newLines = append(newLines, line)
		}
	}

	if found && !exclusive {
		result.Changed = false
		return result, nil
	}

	// Add the key
	newLines = append(newLines, key)
	result.Changed = true

	content := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(authFile, []byte(content), 0600); err != nil {
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}

	// Fix ownership
	fixOwnership(sshDir, user)
	fixOwnership(authFile, user)

	return result, nil
}

// AuthorizedKeyRemove removes a public key from a user's authorized_keys file.
func AuthorizedKeyRemove(user string, key string) (AuthorizedKeyResult, error) {
	result := AuthorizedKeyResult{User: user, Key: key, State: "absent"}

	homeDir, err := userHome(user)
	if err != nil {
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}

	authFile := filepath.Join(homeDir, ".ssh", "authorized_keys")
	data, err := os.ReadFile(authFile)
	if err != nil {
		if os.IsNotExist(err) {
			result.Changed = false
			return result, nil
		}
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}

	keyParts := strings.Fields(key)
	if len(keyParts) < 2 {
		return result, fmt.Errorf("ssh.AuthorizedKey: invalid key format")
	}
	keyType := keyParts[0]
	keyData := keyParts[1]

	lines := strings.Split(string(data), "\n")
	var newLines []string
	removed := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 && fields[0] == keyType && fields[1] == keyData {
			removed = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !removed {
		result.Changed = false
		return result, nil
	}

	content := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(authFile, []byte(content), 0600); err != nil {
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}
	result.Changed = true
	return result, nil
}

// AuthorizedKeyList lists all keys in a user's authorized_keys file.
type AuthorizedKeyListResult struct {
	User string   `json:"user"`
	Keys []string `json:"keys"`
	Count int     `json:"count"`
}

func AuthorizedKeyList(user string) (AuthorizedKeyListResult, error) {
	result := AuthorizedKeyListResult{User: user, Keys: make([]string, 0)}

	homeDir, err := userHome(user)
	if err != nil {
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}

	authFile := filepath.Join(homeDir, ".ssh", "authorized_keys")
	f, err := os.Open(authFile)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, fmt.Errorf("ssh.AuthorizedKey: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			result.Keys = append(result.Keys, line)
		}
	}
	result.Count = len(result.Keys)
	return result, nil
}

func userHome(user string) (string, error) {
	// Try /etc/passwd
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return "", fmt.Errorf("cannot read /etc/passwd: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) >= 6 && parts[0] == user {
			return parts[5], nil
		}
	}
	return "", fmt.Errorf("user %q not found", user)
}

func fixOwnership(path string, user string) {
	// ponytail: best-effort chown; errors silently ignored for non-root execution
	// In production, this would use syscall.Chown with uid/gid lookup
}
