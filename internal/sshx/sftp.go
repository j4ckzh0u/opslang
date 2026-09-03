package sshx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

// SFTPClient provides SFTP file transfer operations.
type SFTPClient struct {
	client *sftp.Client
}

type sftpResult[T any] struct {
	value T
	err   error
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

func runSFTPOperation[T any](s *SFTPClient, ctx context.Context, operation func() (T, error)) (T, error) {
	var zero T
	if ctx == nil {
		return zero, fmt.Errorf("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if s == nil || s.client == nil {
		return zero, fmt.Errorf("SFTP client is not initialized")
	}

	done := make(chan sftpResult[T], 1)
	go func() {
		value, err := operation()
		done <- sftpResult[T]{value: value, err: err}
	}()

	select {
	case result := <-done:
		return result.value, result.err
	case <-ctx.Done():
		closeErr := s.client.Close()
		result := <-done
		if closeErr != nil {
			closeErr = fmt.Errorf("close SFTP client after cancellation: %w", closeErr)
		}
		return zero, errors.Join(ctx.Err(), closeErr, result.err)
	}
}

func closeWithError(current error, closer io.Closer, description string) error {
	if err := closer.Close(); err != nil {
		return errors.Join(current, fmt.Errorf("close %s: %w", description, err))
	}
	return current
}

func validateRemotePath(path, field string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("%s is empty", field)
	}
	return nil
}

// Stat returns metadata for a remote path.
func (s *SFTPClient) Stat(ctx context.Context, remotePath string) (os.FileInfo, error) {
	if err := validateRemotePath(remotePath, "remote path"); err != nil {
		return nil, err
	}
	return runSFTPOperation(s, ctx, func() (os.FileInfo, error) {
		info, statErr := s.client.Stat(remotePath)
		if statErr != nil {
			return nil, fmt.Errorf("stat remote path %s: %w", remotePath, statErr)
		}
		return info, nil
	})
}

// Upload uploads a local file to the remote host.
func (s *SFTPClient) Upload(ctx context.Context, localPath, remotePath string) error {
	_, err := s.UploadAt(ctx, localPath, remotePath, 0)
	return err
}

// UploadAt uploads a local file starting at offset and discards any unconfirmed remote tail.
func (s *SFTPClient) UploadAt(ctx context.Context, localPath, remotePath string, offset int64) (int64, error) {
	return s.uploadAt(ctx, localPath, remotePath, offset, -1)
}

// UploadRangeAt uploads exactly length bytes from offset and discards any unconfirmed remote tail.
func (s *SFTPClient) UploadRangeAt(ctx context.Context, localPath, remotePath string, offset, length int64) (int64, error) {
	if length < 0 {
		return 0, fmt.Errorf("upload length is negative")
	}
	return s.uploadAt(ctx, localPath, remotePath, offset, length)
}

func (s *SFTPClient) uploadAt(ctx context.Context, localPath, remotePath string, offset, length int64) (int64, error) {
	if strings.TrimSpace(localPath) == "" {
		return 0, fmt.Errorf("local path is empty")
	}
	if err := validateRemotePath(remotePath, "remote path"); err != nil {
		return 0, err
	}
	if offset < 0 {
		return 0, fmt.Errorf("upload offset is negative")
	}

	return runSFTPOperation(s, ctx, func() (_ int64, returnErr error) {
		localFile, err := os.Open(localPath)
		if err != nil {
			return 0, fmt.Errorf("open local file %s: %w", localPath, err)
		}
		defer func() {
			returnErr = closeWithError(returnErr, localFile, "local upload file")
		}()
		localInfo, err := localFile.Stat()
		if err != nil {
			return 0, fmt.Errorf("stat local file %s: %w", localPath, err)
		}
		if offset > localInfo.Size() {
			return 0, fmt.Errorf("upload offset %d exceeds local file size %d", offset, localInfo.Size())
		}
		if length >= 0 && length > localInfo.Size()-offset {
			return 0, fmt.Errorf("upload range %d..%d exceeds local file size %d", offset, offset+length, localInfo.Size())
		}

		remoteDir := filepath.Dir(remotePath)
		if err := s.client.MkdirAll(remoteDir); err != nil {
			return 0, fmt.Errorf("create remote directory %s: %w", remoteDir, err)
		}
		flags := os.O_WRONLY | os.O_CREATE
		if offset == 0 {
			flags |= os.O_TRUNC
		} else {
			remoteInfo, err := s.client.Stat(remotePath)
			if err != nil {
				return 0, fmt.Errorf("stat remote file %s for resume: %w", remotePath, err)
			}
			if remoteInfo.Size() < offset {
				return 0, fmt.Errorf("upload offset %d exceeds remote file size %d", offset, remoteInfo.Size())
			}
		}
		remoteFile, err := s.client.OpenFile(remotePath, flags)
		if err != nil {
			return 0, fmt.Errorf("open remote file %s: %w", remotePath, err)
		}
		defer func() {
			returnErr = closeWithError(returnErr, remoteFile, "remote upload file")
		}()
		if offset > 0 {
			if err := remoteFile.Truncate(offset); err != nil {
				return 0, fmt.Errorf("truncate remote file %s to offset %d: %w", remotePath, offset, err)
			}
		}
		if _, err := localFile.Seek(offset, io.SeekStart); err != nil {
			return 0, fmt.Errorf("seek local file %s to offset %d: %w", localPath, offset, err)
		}
		if _, err := remoteFile.Seek(offset, io.SeekStart); err != nil {
			return 0, fmt.Errorf("seek remote file %s to offset %d: %w", remotePath, offset, err)
		}
		var written int64
		if length >= 0 {
			written, err = io.CopyN(remoteFile, localFile, length)
		} else {
			written, err = io.Copy(remoteFile, localFile)
		}
		if err != nil {
			return written, fmt.Errorf("upload file %s at offset %d: %w", remotePath, offset, err)
		}
		return written, nil
	})
}

// Download downloads a file from the remote host.
func (s *SFTPClient) Download(ctx context.Context, remotePath, localPath string) error {
	_, err := s.DownloadAt(ctx, remotePath, localPath, 0)
	return err
}

// DownloadAt downloads a remote file starting at offset and discards any unconfirmed local tail.
func (s *SFTPClient) DownloadAt(ctx context.Context, remotePath, localPath string, offset int64) (int64, error) {
	return s.downloadAt(ctx, remotePath, localPath, offset, -1)
}

// DownloadRangeAt downloads exactly length bytes from offset and discards any unconfirmed local tail.
func (s *SFTPClient) DownloadRangeAt(ctx context.Context, remotePath, localPath string, offset, length int64) (int64, error) {
	if length < 0 {
		return 0, fmt.Errorf("download length is negative")
	}
	return s.downloadAt(ctx, remotePath, localPath, offset, length)
}

func (s *SFTPClient) downloadAt(ctx context.Context, remotePath, localPath string, offset, length int64) (int64, error) {
	if err := validateRemotePath(remotePath, "remote path"); err != nil {
		return 0, err
	}
	if strings.TrimSpace(localPath) == "" {
		return 0, fmt.Errorf("local path is empty")
	}
	if offset < 0 {
		return 0, fmt.Errorf("download offset is negative")
	}

	return runSFTPOperation(s, ctx, func() (_ int64, returnErr error) {
		remoteFile, err := s.client.Open(remotePath)
		if err != nil {
			return 0, fmt.Errorf("open remote file %s: %w", remotePath, err)
		}
		defer func() {
			returnErr = closeWithError(returnErr, remoteFile, "remote download file")
		}()
		remoteInfo, err := remoteFile.Stat()
		if err != nil {
			return 0, fmt.Errorf("stat remote file %s: %w", remotePath, err)
		}
		if offset > remoteInfo.Size() {
			return 0, fmt.Errorf("download offset %d exceeds remote file size %d", offset, remoteInfo.Size())
		}
		if length >= 0 && length > remoteInfo.Size()-offset {
			return 0, fmt.Errorf("download range %d..%d exceeds remote file size %d", offset, offset+length, remoteInfo.Size())
		}

		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return 0, fmt.Errorf("create local directory %s: %w", localDir, err)
		}
		localFile, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return 0, fmt.Errorf("open local file %s: %w", localPath, err)
		}
		defer func() {
			returnErr = closeWithError(returnErr, localFile, "local download file")
		}()
		localInfo, err := localFile.Stat()
		if err != nil {
			return 0, fmt.Errorf("stat local file %s: %w", localPath, err)
		}
		if localInfo.Size() < offset {
			return 0, fmt.Errorf("download offset %d exceeds local file size %d", offset, localInfo.Size())
		}
		if err := localFile.Truncate(offset); err != nil {
			return 0, fmt.Errorf("truncate local file %s to offset %d: %w", localPath, offset, err)
		}
		if _, err := remoteFile.Seek(offset, io.SeekStart); err != nil {
			return 0, fmt.Errorf("seek remote file %s to offset %d: %w", remotePath, offset, err)
		}
		if _, err := localFile.Seek(offset, io.SeekStart); err != nil {
			return 0, fmt.Errorf("seek local file %s to offset %d: %w", localPath, offset, err)
		}
		var written int64
		if length >= 0 {
			written, err = io.CopyN(localFile, remoteFile, length)
		} else {
			written, err = io.Copy(localFile, remoteFile)
		}
		if err != nil {
			return written, fmt.Errorf("download file %s at offset %d: %w", remotePath, offset, err)
		}
		if err := localFile.Sync(); err != nil {
			return 0, fmt.Errorf("sync local file %s: %w", localPath, err)
		}
		return written, nil
	})
}

// Remove removes a remote path.
func (s *SFTPClient) Remove(ctx context.Context, remotePath string) error {
	if err := validateRemotePath(remotePath, "remote path"); err != nil {
		return err
	}
	_, err := runSFTPOperation(s, ctx, func() (struct{}, error) {
		if removeErr := s.client.Remove(remotePath); removeErr != nil {
			return struct{}{}, fmt.Errorf("remove remote path %s: %w", remotePath, removeErr)
		}
		return struct{}{}, nil
	})
	return err
}

// ReadAt reads an exact byte range from a remote file.
func (s *SFTPClient) ReadAt(ctx context.Context, remotePath string, offset int64, length int) ([]byte, error) {
	if err := validateRemotePath(remotePath, "remote path"); err != nil {
		return nil, err
	}
	if offset < 0 {
		return nil, fmt.Errorf("read offset is negative")
	}
	if length < 0 {
		return nil, fmt.Errorf("read length is negative")
	}
	return runSFTPOperation(s, ctx, func() (_ []byte, returnErr error) {
		if length == 0 {
			return []byte{}, nil
		}
		remoteFile, err := s.client.Open(remotePath)
		if err != nil {
			return nil, fmt.Errorf("open remote file %s: %w", remotePath, err)
		}
		defer func() {
			returnErr = closeWithError(returnErr, remoteFile, "remote read file")
		}()
		content := make([]byte, length)
		read, err := remoteFile.ReadAt(content, offset)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("read remote file %s at offset %d: %w", remotePath, offset, err)
		}
		if read != len(content) {
			return nil, fmt.Errorf("read remote file %s at offset %d: %w", remotePath, offset, io.ErrUnexpectedEOF)
		}
		return content, nil
	})
}

// Rename atomically moves a remote file, replacing the destination when supported.
func (s *SFTPClient) Rename(ctx context.Context, oldPath, newPath string) error {
	if err := validateRemotePath(oldPath, "old remote path"); err != nil {
		return err
	}
	if err := validateRemotePath(newPath, "new remote path"); err != nil {
		return err
	}
	_, err := runSFTPOperation(s, ctx, func() (struct{}, error) {
		var renameErr error
		if _, supported := s.client.HasExtension("posix-rename@openssh.com"); supported {
			renameErr = s.client.PosixRename(oldPath, newPath)
		} else {
			renameErr = s.client.Rename(oldPath, newPath)
		}
		if renameErr != nil {
			return struct{}{}, fmt.Errorf("rename remote path %s to %s: %w", oldPath, newPath, renameErr)
		}
		return struct{}{}, nil
	})
	return err
}

// Upload uploads a local file to the remote host via SFTP (convenience method on Client).
func (c *Client) Upload(ctx context.Context, localPath, remotePath string) (returnErr error) {
	sftpClient, err := c.NewSFTPClient()
	if err != nil {
		return err
	}
	defer func() {
		returnErr = closeWithError(returnErr, sftpClient, "SFTP upload client")
	}()

	return sftpClient.Upload(ctx, localPath, remotePath)
}

// Download downloads a file from the remote host via SFTP (convenience method on Client).
func (c *Client) Download(ctx context.Context, remotePath, localPath string) (returnErr error) {
	sftpClient, err := c.NewSFTPClient()
	if err != nil {
		return err
	}
	defer func() {
		returnErr = closeWithError(returnErr, sftpClient, "SFTP download client")
	}()

	return sftpClient.Download(ctx, remotePath, localPath)
}

// RemoteFileChecksum streams the remote file through SHA-256 without
// materializing it locally.
func (s *SFTPClient) RemoteFileChecksum(ctx context.Context, remotePath string) (string, error) {
	if err := validateRemotePath(remotePath, "remote path"); err != nil {
		return "", err
	}
	return runSFTPOperation(s, ctx, func() (_ string, returnErr error) {
		remoteFile, err := s.client.Open(remotePath)
		if err != nil {
			return "", fmt.Errorf("open remote file %s: %w", remotePath, err)
		}
		defer func() {
			returnErr = closeWithError(returnErr, remoteFile, "remote checksum file")
		}()
		h := sha256.New()
		if _, err := io.Copy(h, remoteFile); err != nil {
			return "", fmt.Errorf("read remote file %s: %w", remotePath, err)
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	})
}

// Chmod changes the permissions of a remote file.
func (s *SFTPClient) Chmod(remotePath string, mode os.FileMode) error {
	if err := s.client.Chmod(remotePath, mode); err != nil {
		return fmt.Errorf("failed to chmod remote file %s: %w", remotePath, err)
	}
	return nil
}
