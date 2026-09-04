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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/j4ckzh0u/opslang/internal/sshx"
)

const (
	defaultPartRetention = 24 * time.Hour
	resumeTransferChunk  = 1024 * 1024
	maxMetadataSize      = 1024 * 1024
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
func withSSHClient(ctx context.Context, endpoint string, fn func(c *sshx.Client) error) (returnErr error) {
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
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("close SSH client: %w", closeErr))
		}
	}()
	return fn(client)
}

func effectivePartRetention(retention time.Duration) time.Duration {
	if retention == 0 {
		return defaultPartRetention
	}
	return retention
}

func closeFileWithError(current error, closer io.Closer, description string) error {
	if err := closer.Close(); err != nil {
		return errors.Join(current, fmt.Errorf("close %s: %w", description, err))
	}
	return current
}

func statIfExists(ctx context.Context, sftpClient *sshx.SFTPClient, path string) (os.FileInfo, bool, error) {
	info, err := sftpClient.Stat(ctx, path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return info, true, nil
}

func remoteFileMatches(ctx context.Context, sftpClient *sshx.SFTPClient, path string, size int64, checksum string) (bool, error) {
	info, exists, err := statIfExists(ctx, sftpClient, path)
	if err != nil || !exists {
		return false, err
	}
	if !info.Mode().IsRegular() || info.Size() != size {
		return false, nil
	}
	remoteChecksum, err := sftpClient.RemoteFileChecksum(ctx, path)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(remoteChecksum, checksum), nil
}

func loadRemotePartialMetadata(ctx context.Context, sftpClient *sshx.SFTPClient, path string) (PartialMetadata, error) {
	info, exists, err := statIfExists(ctx, sftpClient, path)
	if err != nil {
		return PartialMetadata{}, err
	}
	if !exists {
		return PartialMetadata{}, os.ErrNotExist
	}
	if info.Size() <= 0 || info.Size() > maxMetadataSize {
		return PartialMetadata{}, fmt.Errorf("remote partial metadata size %d is outside 1..%d", info.Size(), maxMetadataSize)
	}
	content, err := sftpClient.ReadAt(ctx, path, 0, int(info.Size()))
	if err != nil {
		return PartialMetadata{}, err
	}
	var metadata PartialMetadata
	if err := json.Unmarshal(content, &metadata); err != nil {
		return PartialMetadata{}, fmt.Errorf("decode remote partial metadata: %w", err)
	}
	if err := metadata.validateFields(); err != nil {
		return PartialMetadata{}, err
	}
	return metadata, nil
}

func writeRemotePartialMetadata(ctx context.Context, sftpClient *sshx.SFTPClient, path string, metadata PartialMetadata) (returnErr error) {
	if err := metadata.validateFields(); err != nil {
		return err
	}
	content, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode remote partial metadata: %w", err)
	}
	temporary, err := os.CreateTemp("", ".opslang-remote-metadata-*")
	if err != nil {
		return fmt.Errorf("create local metadata staging file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		if removeErr := os.Remove(temporaryPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, fmt.Errorf("remove local metadata staging file: %w", removeErr))
		}
	}()
	if _, err := temporary.Write(content); err != nil {
		return errors.Join(fmt.Errorf("write local metadata staging file: %w", err), temporary.Close())
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close local metadata staging file: %w", err)
	}
	remoteTemporaryPath := path + ".tmp-" + metadata.SessionID
	if err := sftpClient.Upload(ctx, temporaryPath, remoteTemporaryPath); err != nil {
		return fmt.Errorf("upload remote partial metadata: %w", err)
	}
	if err := sftpClient.Rename(ctx, remoteTemporaryPath, path); err != nil {
		return fmt.Errorf("commit remote partial metadata: %w", err)
	}
	return nil
}

func validateRemotePartialPrefix(ctx context.Context, sftpClient *sshx.SFTPClient, sourcePath, partPath string, confirmedSize int64) (returnErr error) {
	if confirmedSize == 0 {
		return nil
	}
	blockSize := int64(partialVerifyBlockSize)
	if confirmedSize < blockSize {
		blockSize = confirmedSize
	}
	offset := confirmedSize - blockSize
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open upload source: %w", err)
	}
	defer func() {
		returnErr = closeFileWithError(returnErr, source, "upload source confirmation file")
	}()
	localBlock := make([]byte, int(blockSize))
	if err := readFullAt(source, localBlock, offset); err != nil {
		return fmt.Errorf("read upload source confirmation block: %w", err)
	}
	remoteBlock, err := sftpClient.ReadAt(ctx, partPath, offset, int(blockSize))
	if err != nil {
		return err
	}
	if !bytes.Equal(localBlock, remoteBlock) {
		return fmt.Errorf("remote partial confirmation block does not match source")
	}
	return nil
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

// SSHResumeUpload uploads through a verified partial file and atomically commits it.
func SSHResumeUpload(ctx context.Context, src, endpoint string, retention time.Duration) (outcome TransferOutcome, returnErr error) {
	_, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return outcome, err
	}
	info, err := os.Stat(src)
	if err != nil {
		return outcome, fmt.Errorf("stat upload source %s: %w", src, err)
	}
	if !info.Mode().IsRegular() {
		return outcome, fmt.Errorf("upload source %s is not a regular file", src)
	}
	checksum, err := computeFileChecksum(src)
	if err != nil {
		return outcome, fmt.Errorf("checksum upload source %s: %w", src, err)
	}
	outcome.Size = info.Size()
	outcome.Checksum = checksum
	outcome.TransferSource = "controller_sftp"

	err = withSSHClient(ctx, endpoint, func(c *sshx.Client) (operationErr error) {
		sftpClient, err := c.NewSFTPClient()
		if err != nil {
			return err
		}
		defer func() {
			operationErr = closeFileWithError(operationErr, sftpClient, "resumable upload SFTP client")
		}()

		matches, err := remoteFileMatches(ctx, sftpClient, remotePath, info.Size(), checksum)
		if err != nil {
			return fmt.Errorf("check remote destination: %w", err)
		}
		if matches {
			outcome.Status = "skipped"
			return nil
		}

		partPath, metadataPath, err := partialPaths(remotePath)
		if err != nil {
			return err
		}
		partInfo, partExists, err := statIfExists(ctx, sftpClient, partPath)
		if err != nil {
			return fmt.Errorf("stat remote partial file: %w", err)
		}
		metadata, metadataErr := loadRemotePartialMetadata(ctx, sftpClient, metadataPath)
		resumeOffset := int64(0)
		canResume := partExists && metadataErr == nil
		if canResume {
			if err := metadata.validateForSource(info.Size(), checksum, partInfo.Size()); err != nil {
				outcome.Warnings = append(outcome.Warnings, "discarding incompatible remote partial state: "+err.Error())
				canResume = false
			} else if metadata.expired(time.Now(), effectivePartRetention(retention)) {
				outcome.Warnings = append(outcome.Warnings, "discarding expired remote partial state")
				canResume = false
			} else if err := validateRemotePartialPrefix(ctx, sftpClient, src, partPath, metadata.ConfirmedSize); err != nil {
				outcome.Warnings = append(outcome.Warnings, "discarding remote partial state with mismatched prefix: "+err.Error())
				canResume = false
			}
		}
		if metadataErr != nil && !errors.Is(metadataErr, os.ErrNotExist) {
			outcome.Warnings = append(outcome.Warnings, "discarding unreadable remote partial metadata: "+metadataErr.Error())
		}
		if canResume {
			resumeOffset = metadata.ConfirmedSize
		} else {
			metadata, err = newPartialMetadata(info.Size(), checksum, time.Now())
			if err != nil {
				return err
			}
		}
		outcome.ResumedBytes = resumeOffset
		if err := writeRemotePartialMetadata(ctx, sftpClient, metadataPath, metadata); err != nil {
			return err
		}

		if info.Size() == 0 {
			if _, err := sftpClient.UploadRangeAt(ctx, src, partPath, 0, 0); err != nil {
				return err
			}
		}
		for offset := resumeOffset; offset < info.Size(); {
			length := min(int64(resumeTransferChunk), info.Size()-offset)
			written, err := sftpClient.UploadRangeAt(ctx, src, partPath, offset, length)
			outcome.TransferredBytes += written
			if err != nil {
				return err
			}
			if written != length {
				return fmt.Errorf("upload wrote %d bytes, expected %d", written, length)
			}
			offset += written
			metadata.ConfirmedSize = offset
			metadata.UpdatedAt = time.Now().UTC()
			if err := writeRemotePartialMetadata(ctx, sftpClient, metadataPath, metadata); err != nil {
				return err
			}
		}
		matches, err = remoteFileMatches(ctx, sftpClient, partPath, info.Size(), checksum)
		if err != nil {
			return fmt.Errorf("verify remote partial file: %w", err)
		}
		if !matches {
			return fmt.Errorf("remote partial file failed final size or SHA-256 verification")
		}
		if err := sftpClient.Rename(ctx, partPath, remotePath); err != nil {
			return err
		}
		if err := sftpClient.Remove(ctx, metadataPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			outcome.Warnings = append(outcome.Warnings, "remove remote partial metadata: "+err.Error())
		}
		outcome.Status = "success"
		outcome.Changed = true
		return nil
	})
	return outcome, err
}

func endpointWithRemotePath(endpoint, remotePath string) (string, error) {
	host, port, user, _, err := parseEndpoint(endpoint)
	if err != nil {
		return "", err
	}
	return formatEndpoint(user, host, port, remotePath), nil
}

// SSHCompressedResumeUpload compresses locally, resumes the compressed object,
// then atomically publishes the original bytes on the remote host.
func SSHCompressedResumeUpload(ctx context.Context, src, endpoint string, retention time.Duration) (outcome TransferOutcome, returnErr error) {
	info, err := os.Stat(src)
	if err != nil {
		return outcome, fmt.Errorf("stat upload source %s: %w", src, err)
	}
	originalChecksum, err := computeFileChecksum(src)
	if err != nil {
		return outcome, fmt.Errorf("checksum upload source %s: %w", src, err)
	}
	compressed, err := gzipFile(src)
	if err != nil {
		return outcome, err
	}
	defer os.Remove(compressed)
	host, port, user, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return outcome, err
	}
	compressedPath := remotePath + ".opslang-compressed"
	compressedEndpoint := formatEndpoint(user, host, port, compressedPath)
	transferred, err := SSHResumeUpload(ctx, compressed, compressedEndpoint, retention)
	if err != nil {
		return outcome, err
	}
	outcome = transferred
	outcome.Size = info.Size()
	outcome.Checksum = originalChecksum
	outcome.TransferSource = "controller_sftp_gzip"
	temporary := remotePath + ".opslang-decompressed"
	err = withSSHClient(ctx, endpoint, func(c *sshx.Client) error {
		command := "gzip -dc -- " + shellQuote(compressedPath) + " > " + shellQuote(temporary)
		result, execErr := c.Exec(ctx, command)
		if execErr != nil {
			return execErr
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("remote decompression failed with exit code %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
		}
		return nil
	})
	if err != nil {
		return outcome, err
	}
	if err := SSHVerifyChecksum(ctx, formatEndpoint(user, host, port, temporary), originalChecksum); err != nil {
		return outcome, fmt.Errorf("verify decompressed upload: %w", err)
	}
	if err := withSSHClient(ctx, endpoint, func(c *sshx.Client) error {
		sftp, sftpErr := c.NewSFTPClient()
		if sftpErr != nil {
			return sftpErr
		}
		defer sftp.Close()
		if renameErr := sftp.Rename(ctx, temporary, remotePath); renameErr != nil {
			return renameErr
		}
		if removeErr := sftp.Remove(ctx, compressedPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return removeErr
		}
		return nil
	}); err != nil {
		return outcome, err
	}
	outcome.Status = "success"
	outcome.Changed = true
	return outcome, nil
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

func localFileMatches(path string, size int64, checksum string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() || info.Size() != size {
		return false, nil
	}
	localChecksum, err := computeFileChecksum(path)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(localChecksum, checksum), nil
}

func validateLocalPartialPrefix(ctx context.Context, sftpClient *sshx.SFTPClient, remotePath, partPath string, confirmedSize int64) (returnErr error) {
	if confirmedSize == 0 {
		return nil
	}
	blockSize := int64(partialVerifyBlockSize)
	if confirmedSize < blockSize {
		blockSize = confirmedSize
	}
	offset := confirmedSize - blockSize
	partial, err := os.Open(partPath)
	if err != nil {
		return fmt.Errorf("open local partial file: %w", err)
	}
	defer func() {
		returnErr = closeFileWithError(returnErr, partial, "local partial confirmation file")
	}()
	localBlock := make([]byte, int(blockSize))
	if err := readFullAt(partial, localBlock, offset); err != nil {
		return fmt.Errorf("read local partial confirmation block: %w", err)
	}
	remoteBlock, err := sftpClient.ReadAt(ctx, remotePath, offset, int(blockSize))
	if err != nil {
		return err
	}
	if !bytes.Equal(localBlock, remoteBlock) {
		return fmt.Errorf("local partial confirmation block does not match remote source")
	}
	return nil
}

// SSHResumeDownload downloads through a verified partial file and atomically commits it.
func SSHResumeDownload(ctx context.Context, endpoint, dst string, retention time.Duration) (outcome TransferOutcome, returnErr error) {
	_, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return outcome, err
	}
	outcome.TransferSource = "controller_sftp"
	err = withSSHClient(ctx, endpoint, func(c *sshx.Client) (operationErr error) {
		sftpClient, err := c.NewSFTPClient()
		if err != nil {
			return err
		}
		defer func() {
			operationErr = closeFileWithError(operationErr, sftpClient, "resumable download SFTP client")
		}()
		remoteInfo, err := sftpClient.Stat(ctx, remotePath)
		if err != nil {
			return fmt.Errorf("stat remote source: %w", err)
		}
		if !remoteInfo.Mode().IsRegular() {
			return fmt.Errorf("remote source %s is not a regular file", remotePath)
		}
		checksum, err := sftpClient.RemoteFileChecksum(ctx, remotePath)
		if err != nil {
			return fmt.Errorf("checksum remote source: %w", err)
		}
		outcome.Size = remoteInfo.Size()
		outcome.Checksum = checksum
		matches, err := localFileMatches(dst, remoteInfo.Size(), checksum)
		if err != nil {
			return fmt.Errorf("check local destination: %w", err)
		}
		if matches {
			outcome.Status = "skipped"
			return nil
		}

		partPath, metadataPath, err := partialPaths(dst)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("create local destination directory: %w", err)
		}
		partInfo, partErr := os.Stat(partPath)
		partExists := partErr == nil
		if partErr != nil && !errors.Is(partErr, os.ErrNotExist) {
			return fmt.Errorf("stat local partial file: %w", partErr)
		}
		metadata, metadataErr := loadPartialMetadata(metadataPath)
		resumeOffset := int64(0)
		canResume := partExists && metadataErr == nil
		if canResume {
			if err := metadata.validateForSource(remoteInfo.Size(), checksum, partInfo.Size()); err != nil {
				outcome.Warnings = append(outcome.Warnings, "discarding incompatible local partial state: "+err.Error())
				canResume = false
			} else if metadata.expired(time.Now(), effectivePartRetention(retention)) {
				outcome.Warnings = append(outcome.Warnings, "discarding expired local partial state")
				canResume = false
			} else if err := validateLocalPartialPrefix(ctx, sftpClient, remotePath, partPath, metadata.ConfirmedSize); err != nil {
				outcome.Warnings = append(outcome.Warnings, "discarding local partial state with mismatched prefix: "+err.Error())
				canResume = false
			}
		}
		if metadataErr != nil && !errors.Is(metadataErr, os.ErrNotExist) {
			outcome.Warnings = append(outcome.Warnings, "discarding unreadable local partial metadata: "+metadataErr.Error())
		}
		if canResume {
			resumeOffset = metadata.ConfirmedSize
		} else {
			metadata, err = newPartialMetadata(remoteInfo.Size(), checksum, time.Now())
			if err != nil {
				return err
			}
		}
		outcome.ResumedBytes = resumeOffset
		if err := writePartialMetadata(metadataPath, metadata); err != nil {
			return err
		}
		if remoteInfo.Size() == 0 {
			if _, err := sftpClient.DownloadRangeAt(ctx, remotePath, partPath, 0, 0); err != nil {
				return err
			}
		}
		for offset := resumeOffset; offset < remoteInfo.Size(); {
			length := min(int64(resumeTransferChunk), remoteInfo.Size()-offset)
			written, err := sftpClient.DownloadRangeAt(ctx, remotePath, partPath, offset, length)
			outcome.TransferredBytes += written
			if err != nil {
				return err
			}
			if written != length {
				return fmt.Errorf("download wrote %d bytes, expected %d", written, length)
			}
			offset += written
			metadata.ConfirmedSize = offset
			metadata.UpdatedAt = time.Now().UTC()
			if err := writePartialMetadata(metadataPath, metadata); err != nil {
				return err
			}
		}
		if err := commitPartialFile(partPath, dst, remoteInfo.Size(), checksum); err != nil {
			return err
		}
		if err := os.Remove(metadataPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			outcome.Warnings = append(outcome.Warnings, "remove local partial metadata: "+err.Error())
		}
		outcome.Status = "success"
		outcome.Changed = true
		return nil
	})
	return outcome, err
}

// SSHCompressedResumeDownload creates a remote gzip stream, resumes it locally,
// then verifies and atomically publishes the decompressed file.
func SSHCompressedResumeDownload(ctx context.Context, endpoint, dst string, retention time.Duration) (outcome TransferOutcome, returnErr error) {
	host, port, user, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return outcome, err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return outcome, fmt.Errorf("create local destination directory: %w", err)
	}
	remoteCompressed := remotePath + ".opslang-compressed"
	compressedEndpoint := formatEndpoint(user, host, port, remoteCompressed)
	if err := withSSHClient(ctx, endpoint, func(c *sshx.Client) error {
		result, execErr := c.Exec(ctx, "gzip -c -- "+shellQuote(remotePath)+" > "+shellQuote(remoteCompressed))
		if execErr != nil {
			return execErr
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("remote compression failed with exit code %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
		}
		return nil
	}); err != nil {
		return outcome, err
	}
	defer func() {
		cleanupErr := withSSHClient(ctx, endpoint, func(c *sshx.Client) error {
			sftp, sftpErr := c.NewSFTPClient()
			if sftpErr != nil {
				return sftpErr
			}
			defer sftp.Close()
			removeErr := sftp.Remove(ctx, remoteCompressed)
			if errors.Is(removeErr, os.ErrNotExist) {
				return nil
			}
			return removeErr
		})
		returnErr = errors.Join(returnErr, cleanupErr)
	}()
	staging, err := os.CreateTemp(filepath.Dir(dst), ".opslang-decompressed-*")
	if err != nil {
		return outcome, fmt.Errorf("create compressed download staging file: %w", err)
	}
	stagingPath := staging.Name()
	_ = staging.Close()
	defer os.Remove(stagingPath)
	compressedLocal := stagingPath + ".gz"
	defer os.Remove(compressedLocal)
	if _, err := SSHResumeDownload(ctx, compressedEndpoint, compressedLocal, retention); err != nil {
		return outcome, err
	}
	if err := gunzipFile(compressedLocal, stagingPath); err != nil {
		return outcome, err
	}
	info, err := os.Stat(stagingPath)
	if err != nil {
		return outcome, err
	}
	checksum, err := computeFileChecksum(stagingPath)
	if err != nil {
		return outcome, err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return outcome, err
	}
	if err := os.Rename(stagingPath, dst); err != nil {
		return outcome, fmt.Errorf("commit decompressed download: %w", err)
	}
	outcome.Status, outcome.Changed = "success", true
	outcome.Size, outcome.Checksum, outcome.TransferSource = info.Size(), checksum, "controller_sftp_gzip"
	return outcome, nil
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

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
	DefaultResumeUploadFunc = SSHResumeUpload
	DefaultResumeDownloadFunc = SSHResumeDownload
	DefaultCompressedResumeUploadFunc = SSHCompressedResumeUpload
	DefaultCompressedResumeDownloadFunc = SSHCompressedResumeDownload
	DefaultVerifyFunc = SSHVerifyChecksum
	DefaultChmodFunc = SSHChmod
	DefaultRelayGroupFunc = SSHRelayGroup
	DefaultCollectRelayGroupFunc = SSHCollectRelayGroup
}
