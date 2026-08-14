package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)

// Upload transfers a local file to a remote path via SFTP over the existing SSH connection.
// ponytail: c.client is thread-safe; Close() during transfer returns errors from sftp ops, not panics.
func (c *Client) Upload(localPath, remotePath string) error {
	if c.IsClosed() {
		return fmt.Errorf("ssh: client is closed")
	}

	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("ssh: sftp client: %w", err)
	}
	defer sftpClient.Close()

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("ssh: open local file %q: %w", localPath, err)
	}
	defer src.Close()

	dst, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("ssh: create remote file %q: %w", remotePath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("ssh: copy to remote %q: %w", remotePath, err)
	}

	return nil
}

// Download transfers a remote file to a local path via SFTP.
func (c *Client) Download(remotePath, localPath string) error {
	if c.IsClosed() {
		return fmt.Errorf("ssh: client is closed")
	}

	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("ssh: sftp client: %w", err)
	}
	defer sftpClient.Close()

	src, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("ssh: open remote file %q: %w", remotePath, err)
	}
	defer src.Close()

	// Ensure local directory exists.
	if dir := filepath.Dir(localPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("ssh: mkdir %q: %w", dir, err)
		}
	}

	dst, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("ssh: create local file %q: %w", localPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("ssh: copy to local %q: %w", localPath, err)
	}

	return nil
}
