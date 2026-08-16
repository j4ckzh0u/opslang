// Host key verification with trust-on-first-use (TOFU) semantics.
package sshx

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// DefaultKnownHostsPath returns the known-hosts file OpsLang maintains.
// Precedence: Config.KnownHostsFile > $OPSLANG_KNOWN_HOSTS >
// ~/.ssh/opslang_known_hosts. OpsLang keeps its own file so it never
// mutates the user's OpenSSH state without consent.
func DefaultKnownHostsPath() string {
	if p := os.Getenv("OPSLANG_KNOWN_HOSTS"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "opslang_known_hosts")
	}
	return filepath.Join(home, ".ssh", "opslang_known_hosts")
}

var knownHostsMu sync.Mutex // serializes appends across concurrent connects

// tofuCallback returns a host key callback implementing TOFU:
//
//   - Known host + matching key  -> accepted
//   - Known host + DIFFERENT key -> rejected (possible MITM)
//   - Unknown host               -> key is appended and accepted
//
// This replaces the previous blanket InsecureIgnoreHostKey, which silently
// disabled all host verification while the project claimed enterprise
// security.
func tofuCallback(path string) (ssh.HostKeyCallback, error) {
	base, err := knownhosts.New(path)
	if err != nil {
		// A missing file is expected on first use; any other parse error
		// must not silently disable verification.
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("hostkey: reading %s: %w", path, err)
		}
		base = func(_ string, _ net.Addr, _ ssh.PublicKey) error { return nil }
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := base(hostname, remote, key)
		if err == nil {
			return nil // known and matching
		}

		var keyErr *knownhosts.KeyError
		if !asKeyError(err, &keyErr) {
			return err // e.g. key mismatch: reject hard
		}
		if len(keyErr.Want) > 0 {
			// The host is known with different keys: possible MITM.
			return fmt.Errorf("hostkey: %s key changed (possible man-in-the-middle); remove the stale entry from %s if this change is intentional", hostname, path)
		}

		// Unknown host: trust on first use - append and accept.
		return appendKnownHost(path, hostname, remote, key)
	}, nil
}

// asKeyError unwraps into *knownhosts.KeyError.
func asKeyError(err error, target **knownhosts.KeyError) bool {
	ke, ok := err.(*knownhosts.KeyError)
	if ok {
		*target = ke
	}
	return ok
}

// appendKnownHost records a newly-seen host key.
func appendKnownHost(path, hostname string, remote net.Addr, key ssh.PublicKey) error {
	knownHostsMu.Lock()
	defer knownHostsMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("hostkey: create dir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("hostkey: open %s: %w", path, err)
	}
	defer f.Close()

	hostport := hostname
	if remote != nil {
		hostport = remote.String()
	}
	line := knownhosts.Line([]string{knownhosts.Normalize(hostport)}, key)
	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("hostkey: write: %w", err)
	}
	return nil
}
