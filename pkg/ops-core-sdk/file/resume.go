package file

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	partialMetadataVersion = 1
	partialFileSuffix      = ".opslang.part"
	partialMetadataSuffix  = ".json"
	partialVerifyBlockSize = 64 * 1024
)

// PartialMetadata identifies the source and the byte range known to be durable.
type PartialMetadata struct {
	Version       int       `json:"version"`
	SessionID     string    `json:"session_id"`
	SourceSize    int64     `json:"source_size"`
	SourceSHA256  string    `json:"source_sha256"`
	ConfirmedSize int64     `json:"confirmed_size"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func partialPaths(finalPath string) (partPath, metadataPath string, err error) {
	if strings.TrimSpace(finalPath) == "" {
		return "", "", fmt.Errorf("final path is empty")
	}
	partPath = finalPath + partialFileSuffix
	return partPath, partPath + partialMetadataSuffix, nil
}

func newPartialMetadata(sourceSize int64, sourceSHA256 string, now time.Time) (PartialMetadata, error) {
	if now.IsZero() {
		return PartialMetadata{}, fmt.Errorf("updated time is zero")
	}
	sessionID, err := newTransferSessionID()
	if err != nil {
		return PartialMetadata{}, err
	}
	metadata := PartialMetadata{
		Version:      partialMetadataVersion,
		SessionID:    sessionID,
		SourceSize:   sourceSize,
		SourceSHA256: strings.ToLower(sourceSHA256),
		UpdatedAt:    now.UTC(),
	}
	if err := metadata.validateFields(); err != nil {
		return PartialMetadata{}, err
	}
	return metadata, nil
}

func newTransferSessionID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate transfer session ID: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}

func (m PartialMetadata) validateFields() error {
	if m.Version != partialMetadataVersion {
		return fmt.Errorf("unsupported partial metadata version %d", m.Version)
	}
	if strings.TrimSpace(m.SessionID) == "" {
		return fmt.Errorf("partial metadata session_id is empty")
	}
	if m.SourceSize < 0 {
		return fmt.Errorf("partial metadata source_size is negative")
	}
	if !validSHA256(m.SourceSHA256) {
		return fmt.Errorf("partial metadata source_sha256 must be 64 hexadecimal characters")
	}
	if m.ConfirmedSize < 0 || m.ConfirmedSize > m.SourceSize {
		return fmt.Errorf("partial metadata confirmed_size %d is outside source range 0..%d", m.ConfirmedSize, m.SourceSize)
	}
	if m.UpdatedAt.IsZero() {
		return fmt.Errorf("partial metadata updated_at is zero")
	}
	return nil
}

func (m PartialMetadata) validateForSource(sourceSize int64, sourceSHA256 string, partSize int64) error {
	if err := m.validateFields(); err != nil {
		return err
	}
	if sourceSize < 0 {
		return fmt.Errorf("source size is negative")
	}
	if partSize < 0 {
		return fmt.Errorf("partial file size is negative")
	}
	if m.SourceSize != sourceSize {
		return fmt.Errorf("partial metadata source size %d does not match %d", m.SourceSize, sourceSize)
	}
	if !strings.EqualFold(m.SourceSHA256, sourceSHA256) {
		return fmt.Errorf("partial metadata source checksum does not match")
	}
	if partSize < m.ConfirmedSize || partSize > sourceSize {
		return fmt.Errorf("partial file size %d is incompatible with confirmed size %d and source size %d", partSize, m.ConfirmedSize, sourceSize)
	}
	return nil
}

func (m PartialMetadata) expired(now time.Time, retention time.Duration) bool {
	if retention <= 0 || now.Before(m.UpdatedAt) {
		return false
	}
	return now.Sub(m.UpdatedAt) >= retention
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validatePartialPrefix(source, partial io.ReaderAt, confirmedSize int64) error {
	if source == nil || partial == nil {
		return fmt.Errorf("source and partial readers are required")
	}
	if confirmedSize < 0 {
		return fmt.Errorf("confirmed size is negative")
	}
	if confirmedSize == 0 {
		return nil
	}
	blockSize := int64(partialVerifyBlockSize)
	if confirmedSize < blockSize {
		blockSize = confirmedSize
	}
	offset := confirmedSize - blockSize
	sourceBlock := make([]byte, int(blockSize))
	partialBlock := make([]byte, int(blockSize))
	if err := readFullAt(source, sourceBlock, offset); err != nil {
		return fmt.Errorf("read source confirmation block: %w", err)
	}
	if err := readFullAt(partial, partialBlock, offset); err != nil {
		return fmt.Errorf("read partial confirmation block: %w", err)
	}
	if !bytes.Equal(sourceBlock, partialBlock) {
		return fmt.Errorf("partial file confirmation block does not match source")
	}
	return nil
}

func readFullAt(reader io.ReaderAt, destination []byte, offset int64) error {
	read, err := reader.ReadAt(destination, offset)
	if err != nil && !(err == io.EOF && read == len(destination)) {
		return err
	}
	if read != len(destination) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func loadPartialMetadata(path string) (PartialMetadata, error) {
	if strings.TrimSpace(path) == "" {
		return PartialMetadata{}, fmt.Errorf("partial metadata path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return PartialMetadata{}, fmt.Errorf("read partial metadata %s: %w", path, err)
	}
	var metadata PartialMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return PartialMetadata{}, fmt.Errorf("decode partial metadata %s: %w", path, err)
	}
	if err := metadata.validateFields(); err != nil {
		return PartialMetadata{}, fmt.Errorf("validate partial metadata %s: %w", path, err)
	}
	return metadata, nil
}

func writePartialMetadata(path string, metadata PartialMetadata) (returnErr error) {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("partial metadata path is empty")
	}
	if err := metadata.validateFields(); err != nil {
		return err
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode partial metadata: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".opslang-metadata-*")
	if err != nil {
		return fmt.Errorf("create temporary partial metadata: %w", err)
	}
	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		if !committed {
			if err := os.Remove(temporaryPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				returnErr = errors.Join(returnErr, fmt.Errorf("remove temporary partial metadata: %w", err))
			}
		}
	}()

	if err := temporary.Chmod(0600); err != nil {
		return errors.Join(fmt.Errorf("set temporary partial metadata mode: %w", err), temporary.Close())
	}
	if _, err := temporary.Write(data); err != nil {
		return errors.Join(fmt.Errorf("write temporary partial metadata: %w", err), temporary.Close())
	}
	if err := temporary.Sync(); err != nil {
		return errors.Join(fmt.Errorf("sync temporary partial metadata: %w", err), temporary.Close())
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary partial metadata: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("commit partial metadata %s: %w", path, err)
	}
	committed = true
	return nil
}

func commitPartialFile(partPath, finalPath string, expectedSize int64, expectedSHA256 string) error {
	if strings.TrimSpace(partPath) == "" || strings.TrimSpace(finalPath) == "" {
		return fmt.Errorf("partial and final paths are required")
	}
	if expectedSize < 0 {
		return fmt.Errorf("expected size is negative")
	}
	if !validSHA256(expectedSHA256) {
		return fmt.Errorf("expected SHA-256 must be 64 hexadecimal characters")
	}
	info, err := os.Lstat(partPath)
	if err != nil {
		return fmt.Errorf("stat partial file %s: %w", partPath, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("partial file %s is not a regular file", partPath)
	}
	if info.Size() != expectedSize {
		return fmt.Errorf("partial file size %d does not match expected %d", info.Size(), expectedSize)
	}
	checksum, err := computeFileChecksum(partPath)
	if err != nil {
		return fmt.Errorf("checksum partial file %s: %w", partPath, err)
	}
	if !strings.EqualFold(checksum, expectedSHA256) {
		return fmt.Errorf("partial file checksum does not match expected SHA-256")
	}
	if err := os.Rename(partPath, finalPath); err != nil {
		return fmt.Errorf("commit partial file %s: %w", finalPath, err)
	}
	return nil
}
