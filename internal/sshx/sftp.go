package sshx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)

// SFTPClient provides SFTP file transfer operations.
type SFTPClient struct {
	client *sftp.Client
}

// NewSFTPClient creates a new SFTP client from an SSH connection.
func (c *Client) NewSFTPClient() (*SFTPClient, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	sftpClient, err := c.sftpNewClient(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &SFTPClient{client: sftpClient}, nil
}

// Close closes the SFTP client connection.
func (s *SFTPClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Upload uploads a local file to the remote host.
func (s *SFTPClient) Upload(ctx context.Context, localPath, remotePath string) error {
	// Ensure remote directory exists.
	remoteDir := filepath.Dir(remotePath)
	if err := s.client.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory %s: %w", remoteDir, err)
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	remoteFile, err := s.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(remoteFile, localFile)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}
		return nil
	}
}

// Download downloads a file from the remote host.
func (s *SFTPClient) Download(ctx context.Context, remotePath, localPath string) error {
	remoteFile, err := s.client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory %s: %w", localDir, err)
	}

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(localFile, remoteFile)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}
		return nil
	}
}

// Upload uploads a local file to the remote host via SFTP (convenience method on Client).
func (c *Client) Upload(ctx context.Context, localPath, remotePath string) error {
	sftpClient, err := c.NewSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return sftpClient.Upload(ctx, localPath, remotePath)
}

// Download downloads a file from the remote host via SFTP (convenience method on Client).
func (c *Client) Download(ctx context.Context, remotePath, localPath string) error {
	sftpClient, err := c.NewSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return sftpClient.Download(ctx, remotePath, localPath)
}

// RemoteFileChecksum streams the remote file through SHA-256 without
// materializing it locally.
func (s *SFTPClient) RemoteFileChecksum(ctx context.Context, remotePath string) (string, error) {
	remoteFile, err := s.client.Open(remotePath)
	if err != nil {
		return "", fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	h := sha256.New()
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(h, remoteFile)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("failed to read remote file %s: %w", remotePath, err)
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Chmod changes the permissions of a remote file.
func (s *SFTPClient) Chmod(remotePath string, mode os.FileMode) error {
	if err := s.client.Chmod(remotePath, mode); err != nil {
		return fmt.Errorf("failed to chmod remote file %s: %w", remotePath, err)
	}
	return nil
}
