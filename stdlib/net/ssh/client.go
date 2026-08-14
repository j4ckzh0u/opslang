package ssh

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// ExecResult holds the output of a remote command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// DefaultCommandTimeout is the default per-command execution timeout.
const DefaultCommandTimeout = 30 * time.Second

// safeBuffer is a concurrency-safe bytes.Buffer wrapper.
// Used when multiple goroutines (stdout + stderr copy) write to the same buffer.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *safeBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *safeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

// Client wraps a single SSH connection to a remote host.
type Client struct {
	cfg    Config
	client *gossh.Client
	mu     sync.Mutex
	closed bool
}

// NewClient establishes an SSH connection using the provided Config.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("ssh: host is required")
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("ssh: user is required")
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultPort
	}

	sshCfg, err := buildSSHConfig(cfg)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := gossh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("ssh: dial %s: %w", addr, err)
	}

	return &Client{cfg: cfg, client: conn}, nil
}

// Run executes cmd on the remote host and returns separated stdout, stderr, and exit code.
// It uses explicit pipes with goroutines to ensure stdout/stderr are fully drained
// even when the command exits abnormally (e.g. SIGPIPE from pipe commands like "cmd | head").
func (c *Client) Run(cmd string) (*ExecResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("ssh: client is closed")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh: new session: %w", err)
	}
	defer session.Close()

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ssh: stdout pipe: %w", err)
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("ssh: stderr pipe: %w", err)
	}

	var stdout, stderr bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(&stdout, stdoutPipe)
	}()
	go func() {
		defer wg.Done()
		io.Copy(&stderr, stderrPipe)
	}()

	exitCode := 0
	startErr := session.Start(cmd)
	if startErr != nil {
		// Start failed — wait for goroutines to finish (they'll see closed pipes),
		// then return whatever partial output was captured.
		wg.Wait()
		return &ExecResult{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: 255,
		}, startErr
	}

	// Timeout: signal the session if it takes too long.
	timeout := DefaultCommandTimeout
	if c.cfg.Timeout > 0 {
		timeout = c.cfg.Timeout
	}
	timer := time.AfterFunc(timeout, func() {
		session.Signal(gossh.SIGKILL)
	})

	waitErr := session.Wait()
	timer.Stop()
	// Wait for all output to be fully drained before reading buffers.
	wg.Wait()

	if waitErr != nil {
		if exitErr, ok := waitErr.(*gossh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			return &ExecResult{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				ExitCode: 255,
			}, waitErr
		}
	}

	return &ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

// CombinedOutput executes cmd and returns combined stdout+stderr.
// Like Run, it uses explicit pipes to ensure complete output capture
// even when commands exit abnormally (e.g. SIGPIPE, exit 141).
func (c *Client) CombinedOutput(cmd string) (string, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return "", 255, fmt.Errorf("ssh: client is closed")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", 255, fmt.Errorf("ssh: new session: %w", err)
	}
	defer session.Close()

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return "", 255, fmt.Errorf("ssh: stdout pipe: %w", err)
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return "", 255, fmt.Errorf("ssh: stderr pipe: %w", err)
	}

	var combined safeBuffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(&combined, stdoutPipe)
	}()
	go func() {
		defer wg.Done()
		io.Copy(&combined, stderrPipe)
	}()

	exitCode := 0
	startErr := session.Start(cmd)
	if startErr != nil {
		wg.Wait()
		return combined.String(), 255, startErr
	}

	// Timeout: signal the session if it takes too long.
	timeout := DefaultCommandTimeout
	if c.cfg.Timeout > 0 {
		timeout = c.cfg.Timeout
	}
	timer := time.AfterFunc(timeout, func() {
		session.Signal(gossh.SIGKILL)
	})

	waitErr := session.Wait()
	timer.Stop()
	wg.Wait()

	if waitErr != nil {
		if exitErr, ok := waitErr.(*gossh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			return combined.String(), 255, waitErr
		}
	}

	return combined.String(), exitCode, nil
}

// NewSession creates a new SSH session (caller must close it).
// Useful for SFTP or interactive use.
func (c *Client) NewSession() (*gossh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("ssh: client is closed")
	}
	return c.client.NewSession()
}

// Close closes the underlying SSH connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	return c.client.Close()
}

// IsClosed reports whether the client has been closed.
func (c *Client) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// Underlying returns the raw *ssh.Client for advanced use (e.g. SFTP).
func (c *Client) Underlying() *gossh.Client {
	return c.client
}

// keepAlive starts a heartbeat on the connection. Not used by default but
// available for long-running pool connections.
func (c *Client) keepAlive(interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return
			}
			_, _, err := c.client.SendRequest("keepalive@openssh.com", true, nil)
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}
