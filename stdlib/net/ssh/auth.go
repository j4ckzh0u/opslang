// Package ssh provides a native Go SSH client implementation,
// replacing external sshpass/shell command invocations.
package ssh

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config holds SSH connection parameters.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	KeyFile  string
	Timeout  time.Duration
}

// DefaultPort is used when Config.Port is zero.
const DefaultPort = 22

// DefaultTimeout is used when Config.Timeout is zero.
const DefaultTimeout = 10 * time.Second

// buildSSHConfig constructs an *ssh.ClientConfig from the given Config.
// Authentication methods are tried in order: key-based first, then password.
func buildSSHConfig(cfg Config) (*ssh.ClientConfig, error) {
	var methods []ssh.AuthMethod

	// Key-based authentication (higher priority).
	if cfg.KeyFile != "" {
		method, err := keyAuthMethod(cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("ssh: failed to load key file %q: %w", cfg.KeyFile, err)
		}
		methods = append(methods, method)
	}

	// Password authentication.
	if cfg.Password != "" {
		methods = append(methods, ssh.Password(cfg.Password))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("ssh: no authentication method provided (need password or key file)")
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	return &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            methods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // ops tooling: matches prior ssh -o StrictHostKeyChecking=no
		Timeout:         timeout,
	}, nil
}

// keyAuthMethod reads a PEM-encoded private key from path and returns an AuthMethod.
func keyAuthMethod(path string) (ssh.AuthMethod, error) {
	pem, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}
	return ssh.PublicKeys(signer), nil
}
