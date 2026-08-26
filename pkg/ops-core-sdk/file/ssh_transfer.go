// Real SSH/SFTP implementations of the transfer hooks used by Distribute
// and Collect. Credentials come from the environment so they never appear
// in scripts:
//
//	OPSLANG_SSH_USER     default remote user (fallback: root)
//	OPSLANG_SSH_PASSWORD password auth
//	OPSLANG_SSH_KEY      private key path
//
// A target may override user/port via its DistributeTarget/CollectTarget
// fields; the endpoint string format is ssh://user@host:port/remote/path.
package file

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/j4ckzh0u/opslang/internal/sshx"
)

// formatEndpoint renders a remote endpoint as ssh://user@host:port/remote/path.
func formatEndpoint(user, host string, port int, remotePath string) string {
	u := url.URL{
		Scheme: "ssh",
		User:   url.User(user),
		Host:   host + ":" + strconv.Itoa(port),
		Path:   remotePath,
	}
	return u.String()
}

// parseEndpoint splits an ssh:// endpoint back into its parts.
func parseEndpoint(endpoint string) (host string, port int, user string, remotePath string, err error) {
	u, perr := url.Parse(endpoint)
	if perr != nil || u.Scheme != "ssh" {
		return "", 0, "", "", fmt.Errorf("invalid ssh endpoint %q (want ssh://user@host:port/path)", endpoint)
	}
	host = u.Hostname()
	port = 22
	if p := u.Port(); p != "" {
		if v, e := strconv.Atoi(p); e == nil {
			port = v
		}
	}
	user = "root"
	if u.User != nil && u.User.Username() != "" {
		user = u.User.Username()
	}
	remotePath = u.Path
	return host, port, user, remotePath, nil
}

// sshAuth resolves credentials from the environment. Password lookup is
// per-host first (OPSLANG_SSH_PASSWORD_<HOST with dots as underscores>),
// then global (OPSLANG_SSH_PASSWORD) - heterogeneous labs have different
// passwords per machine while scripts stay credential-free.
func sshAuth(host string) (password, keyFile string) {
	if host != "" {
		specific := "OPSLANG_SSH_PASSWORD_" + strings.ToUpper(strings.ReplaceAll(host, ".", "_"))
		if p := os.Getenv(specific); p != "" {
			password = p
		}
	}
	if password == "" {
		password = os.Getenv("OPSLANG_SSH_PASSWORD")
	}
	return password, os.Getenv("OPSLANG_SSH_KEY")
}

// withSSHClient dials the endpoint and runs fn with the connected client.
func withSSHClient(ctx context.Context, endpoint string, fn func(c *sshx.Client) error) error {
	host, port, user, _, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	password, keyFile := sshAuth(host)

	cfg := &sshx.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		KeyFile:  keyFile,
		Timeout:  30 * time.Second,
		Retries:  1,
	}
	client, err := sshx.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("ssh dial %s: %w", host, err)
	}
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("ssh connect %s: %w", host, err)
	}
	defer client.Close()
	return fn(client)
}

// SSHTransfer uploads a local file to a remote endpoint (ssh://...).
// It is the default TransferFunc for Distribute.
func SSHTransfer(ctx context.Context, src, endpoint string) error {
	_, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	return withSSHClient(ctx, endpoint, func(c *sshx.Client) error {
		return c.Upload(ctx, src, remotePath)
	})
}

// SSHCollectDownload downloads a remote file to a local path.
// It is the default download function for Collect.
func SSHCollectDownload(ctx context.Context, endpoint, dst string) error {
	_, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	return withSSHClient(ctx, endpoint, func(c *sshx.Client) error {
		return c.Download(ctx, remotePath, dst)
	})
}

// SSHVerifyChecksum streams the remote file through SHA-256 and compares
// it with the expected digest. It is the default verification hook.
func SSHVerifyChecksum(ctx context.Context, endpoint, wantSHA256 string) error {
	_, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	return withSSHClient(ctx, endpoint, func(c *sshx.Client) error {
		sftp, err := c.NewSFTPClient()
		if err != nil {
			return err
		}
		defer sftp.Close()
		got, err := sftp.RemoteFileChecksum(ctx, remotePath)
		if err != nil {
			return err
		}
		if !strings.EqualFold(got, wantSHA256) {
			return fmt.Errorf("checksum mismatch for %s: remote=%s want=%s", remotePath, got, wantSHA256)
		}
		return nil
	})
}

// SSHChmod applies an octal mode string to the remote file. It is the
// default mode hook for Distribute.
func SSHChmod(ctx context.Context, endpoint, mode string) error {
	_, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	var m uint64
	if _, err := fmt.Sscanf(mode, "%o", &m); err != nil {
		return fmt.Errorf("invalid mode %q (want octal like 0644)", mode)
	}
	return withSSHClient(ctx, endpoint, func(c *sshx.Client) error {
		sftp, err := c.NewSFTPClient()
		if err != nil {
			return err
		}
		defer sftp.Close()
		return sftp.Chmod(remotePath, os.FileMode(m))
	})
}

// WireSSHTransfer installs the real SSH/SFTP implementations as the
// package defaults. opsctl calls this at startup; tests and embeddings
// that want different behavior can install their own hooks instead.
func WireSSHTransfer() {
	DefaultTransferFunc = SSHTransfer
	DefaultCollectDownloadFunc = SSHCollectDownload
	DefaultVerifyFunc = SSHVerifyChecksum
	DefaultChmodFunc = SSHChmod
}
