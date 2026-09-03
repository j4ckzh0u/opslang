package sshx

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// DialFunc is a function type for establishing SSH connections.
// It can be injected for testing purposes.
type DialFunc func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)

// SFTPClientFactory creates an SFTP client from an SSH connection.
// Can be overridden in tests.
type SFTPClientFactory func(conn *ssh.Client) (*sftp.Client, error)

// Client wraps an SSH connection with retry, timeout, and execution capabilities.
type Client struct {
	config        *Config
	sshConfig     *ssh.ClientConfig
	dial          DialFunc
	sftpNewClient SFTPClientFactory
	mu            sync.Mutex
	conn          *ssh.Client
}

// NewClient creates a new SSH client with the given configuration.
func NewClient(cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cfg.SetDefaults()

	sshConfig, err := buildSSHConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH config: %w", err)
	}

	return &Client{
		config:        cfg,
		sshConfig:     sshConfig,
		dial:          ssh.Dial,
		sftpNewClient: defaultSFTPClientFactory,
	}, nil
}

// defaultSFTPClientFactory is the default SFTP client factory.
func defaultSFTPClientFactory(conn *ssh.Client) (*sftp.Client, error) {
	return sftp.NewClient(conn)
}

// buildSSHConfig constructs an ssh.ClientConfig from the provided configuration.
func buildSSHConfig(cfg *Config) (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod

	// Add key-based authentication if KeyFile is provided.
	if cfg.KeyFile != "" {
		keyData, err := os.ReadFile(cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file %s: %w", cfg.KeyFile, err)
		}

		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key %s: %w", cfg.KeyFile, err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Add password authentication if Password is provided.
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided (need password or key file)")
	}

	hostKeyCB := ssh.HostKeyCallback(ssh.InsecureIgnoreHostKey())
	if !cfg.InsecureSkipHostKeyVerify {
		knownHostsPath := cfg.KnownHostsFile
		if knownHostsPath == "" {
			knownHostsPath = DefaultKnownHostsPath()
		}
		cb, err := tofuCallback(knownHostsPath)
		if err != nil {
			return nil, err
		}
		hostKeyCB = cb
	}

	return &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		Timeout:         cfg.Timeout,
		HostKeyCallback: hostKeyCB,
	}, nil
}

// Connect establishes an SSH connection with retry logic.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	var lastErr error
	for attempt := 0; attempt <= c.config.Retries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, err := c.dial("tcp", addr, c.sshConfig)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt+1, err)
			if attempt < c.config.Retries {
				// Exponential backoff
				backoff := time.Duration(1<<uint(attempt)) * time.Second
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					continue
				}
			}
			continue
		}

		c.conn = conn
		return nil
	}

	return fmt.Errorf("failed to connect to %s after %d attempts: %w", addr, c.config.Retries+1, lastErr)
}

// Close closes the SSH connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// IsConnected returns true if the client has an active connection.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

// ExecResult holds the result of executing a remote command.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// Exec executes a command on the remote host with context-based timeout.
func (c *Client) Exec(ctx context.Context, cmd string) (*ExecResult, error) {
	return c.ExecWithStdin(ctx, cmd, nil)
}

// ExecWithStdin executes a command on the remote host with stdin data.
// If stdinData is nil, an empty stdin is provided to the command.
func (c *Client) ExecWithStdin(ctx context.Context, cmd string, stdinData []byte) (*ExecResult, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	session, err := conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set stdin from provided data.
	session.Stdin = bytes.NewReader(stdinData)

	var stdout, stderr safeBuffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- session.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return nil, ctx.Err()
	case err := <-done:
		result := &ExecResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}

		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				result.ExitCode = exitErr.ExitStatus()
				return result, nil
			}
			return nil, fmt.Errorf("command execution failed: %w", err)
		}

		return result, nil
	}
}

// safeBuffer is a thread-safe buffer for capturing stdout/stderr.
type safeBuffer struct {
	mu  sync.Mutex
	buf []byte
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}
