package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// RelayFetchOptions describes an authenticated, pinned relay download.
type RelayFetchOptions struct {
	URL             string
	Token           string
	CertFingerprint string
	SHA256          string
	Size            int64
	Dest            string
	PartRetention   time.Duration
	Timeout         time.Duration
}

// RelayFetch downloads one file with Range resume and atomic replacement.
func RelayFetch(ctx context.Context, opts RelayFetchOptions) (outcome TransferOutcome, returnErr error) {
	if ctx == nil {
		return TransferOutcome{}, fmt.Errorf("relay fetch context is nil")
	}
	if strings.TrimSpace(opts.URL) == "" || strings.TrimSpace(opts.Token) == "" || strings.TrimSpace(opts.Dest) == "" {
		return TransferOutcome{}, fmt.Errorf("relay fetch URL, token, and destination are required")
	}
	if !validSHA256(opts.CertFingerprint) || !validSHA256(opts.SHA256) {
		return TransferOutcome{}, fmt.Errorf("relay certificate fingerprint and source SHA-256 must be 64 hexadecimal characters")
	}
	if opts.Size < 0 {
		return TransferOutcome{}, fmt.Errorf("relay source size is negative")
	}
	if opts.PartRetention < 0 || opts.Timeout < 0 {
		return TransferOutcome{}, fmt.Errorf("relay retention and timeout must not be negative")
	}
	outcome = TransferOutcome{Checksum: strings.ToLower(opts.SHA256), Size: opts.Size, TransferSource: "relay"}
	if match, err := localFileMatches(opts.Dest, opts.Size, opts.SHA256); err != nil {
		return outcome, err
	} else if match {
		outcome.Status = "skipped"
		return outcome, nil
	}
	partPath, metadataPath, err := partialPaths(opts.Dest)
	if err != nil {
		return outcome, err
	}
	if err := os.MkdirAll(filepath.Dir(opts.Dest), 0o755); err != nil {
		return outcome, fmt.Errorf("create relay destination directory: %w", err)
	}
	client := relayHTTPClient(opts.CertFingerprint, opts.Timeout)
	retention := effectivePartRetention(opts.PartRetention)
	metadata, offset, warnings, err := prepareRelayPartial(ctx, client, opts, partPath, metadataPath, retention)
	outcome.Warnings = append(outcome.Warnings, warnings...)
	if err != nil {
		return outcome, err
	}
	outcome.ResumedBytes = offset

	part, err := os.OpenFile(partPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return outcome, fmt.Errorf("open relay partial file: %w", err)
	}
	defer func() {
		if part != nil {
			returnErr = closeFileWithError(returnErr, part, "relay partial file")
		}
	}()
	if err := part.Truncate(offset); err != nil {
		return outcome, fmt.Errorf("truncate relay partial file: %w", err)
	}
	if _, err := part.Seek(offset, io.SeekStart); err != nil {
		return outcome, fmt.Errorf("seek relay partial file: %w", err)
	}

	response, err := relayRequest(ctx, client, opts, offset, -1)
	if err != nil {
		return outcome, err
	}
	defer func() { returnErr = closeFileWithError(returnErr, response.Body, "relay response body") }()
	written, err := copyRelayResponse(part, response.Body, opts.Size-offset, func(delta int64) error {
		metadata.ConfirmedSize += delta
		metadata.UpdatedAt = time.Now().UTC()
		return writePartialMetadata(metadataPath, metadata)
	})
	outcome.TransferredBytes = written
	if err != nil {
		return outcome, err
	}
	if err := part.Sync(); err != nil {
		return outcome, fmt.Errorf("sync relay partial file: %w", err)
	}
	if err := part.Close(); err != nil {
		return outcome, fmt.Errorf("close relay partial file before commit: %w", err)
	}
	part = nil
	if err := commitPartialFile(partPath, opts.Dest, opts.Size, opts.SHA256); err != nil {
		return outcome, err
	}
	if err := os.Remove(metadataPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		outcome.Warnings = append(outcome.Warnings, "remove relay partial metadata: "+err.Error())
	}
	outcome.Status = "success"
	outcome.Changed = true
	return outcome, nil
}

func relayHTTPClient(fingerprint string, timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // Exact leaf pinning replaces CA and hostname verification for ephemeral certificates.
		VerifyConnection: func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) == 0 {
				return fmt.Errorf("relay server did not provide a certificate")
			}
			digest := sha256.Sum256(state.PeerCertificates[0].Raw)
			if !strings.EqualFold(hex.EncodeToString(digest[:]), fingerprint) {
				return fmt.Errorf("relay certificate fingerprint mismatch")
			}
			return nil
		},
	}}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return fmt.Errorf("relay redirects are refused")
		},
	}
}

func prepareRelayPartial(ctx context.Context, client *http.Client, opts RelayFetchOptions, partPath, metadataPath string, retention time.Duration) (PartialMetadata, int64, []string, error) {
	reset := func(reason string) (PartialMetadata, int64, []string, error) {
		metadata, err := newPartialMetadata(opts.Size, opts.SHA256, time.Now().UTC())
		if err != nil {
			return PartialMetadata{}, 0, nil, err
		}
		if err := writePartialMetadata(metadataPath, metadata); err != nil {
			return PartialMetadata{}, 0, nil, err
		}
		warnings := []string(nil)
		if reason != "" {
			warnings = append(warnings, reason+"; restarting relay transfer")
		}
		return metadata, 0, warnings, nil
	}
	metadata, err := loadPartialMetadata(metadataPath)
	if errors.Is(err, os.ErrNotExist) {
		return reset("")
	}
	if err != nil {
		return reset(err.Error())
	}
	partInfo, err := os.Stat(partPath)
	if err != nil {
		return reset("partial file is unavailable")
	}
	if err := metadata.validateForSource(opts.Size, opts.SHA256, partInfo.Size()); err != nil || metadata.expired(time.Now().UTC(), retention) {
		if err != nil {
			return reset(err.Error())
		}
		return reset("partial metadata expired")
	}
	if metadata.ConfirmedSize > 0 {
		blockSize := min(int64(partialVerifyBlockSize), metadata.ConfirmedSize)
		start := metadata.ConfirmedSize - blockSize
		response, requestErr := relayRequest(ctx, client, opts, start, metadata.ConfirmedSize-1)
		if requestErr != nil {
			return PartialMetadata{}, 0, nil, requestErr
		}
		remote, readErr := io.ReadAll(io.LimitReader(response.Body, blockSize+1))
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil || int64(len(remote)) != blockSize {
			return reset("relay confirmation block could not be read")
		}
		local := make([]byte, blockSize)
		part, openErr := os.Open(partPath)
		if openErr != nil {
			return reset("partial file could not be opened")
		}
		readErr = readFullAt(part, local, start)
		closeErr = part.Close()
		if readErr != nil || closeErr != nil || !bytes.Equal(local, remote) {
			return reset("partial confirmation block does not match relay source")
		}
	}
	return metadata, metadata.ConfirmedSize, nil, nil
}

func relayRequest(ctx context.Context, client *http.Client, opts RelayFetchOptions, start, end int64) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create relay request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+opts.Token)
	if start > 0 || end >= 0 {
		value := "bytes=" + strconv.FormatInt(start, 10) + "-"
		if end >= 0 {
			value += strconv.FormatInt(end, 10)
		}
		request.Header.Set("Range", value)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch relay file: %w", err)
	}
	wantStatus := http.StatusOK
	if start > 0 || end >= 0 {
		wantStatus = http.StatusPartialContent
	}
	if response.StatusCode != wantStatus {
		response.Body.Close()
		return nil, fmt.Errorf("fetch relay file: HTTP status %d", response.StatusCode)
	}
	if checksum := response.Header.Get("X-Content-SHA256"); !strings.EqualFold(checksum, opts.SHA256) {
		response.Body.Close()
		return nil, fmt.Errorf("relay source checksum header does not match")
	}
	return response, nil
}

func copyRelayResponse(destination io.Writer, source io.Reader, expected int64, checkpoint func(int64) error) (int64, error) {
	if expected < 0 {
		return 0, fmt.Errorf("expected relay byte count is negative")
	}
	buffer := make([]byte, resumeTransferChunk)
	var total int64
	for total < expected {
		limit := min(int64(len(buffer)), expected-total)
		read, readErr := io.ReadFull(source, buffer[:limit])
		if read > 0 {
			written, writeErr := destination.Write(buffer[:read])
			total += int64(written)
			if writeErr != nil {
				return total, fmt.Errorf("write relay partial file: %w", writeErr)
			}
			if written != read {
				return total, io.ErrShortWrite
			}
			if err := checkpoint(int64(written)); err != nil {
				return total, fmt.Errorf("checkpoint relay transfer: %w", err)
			}
		}
		if readErr != nil {
			return total, fmt.Errorf("read relay response: %w", readErr)
		}
	}
	return total, nil
}
