package relay

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
)

// ChunkedTransfer handles file transfers with deduplication via content hashing.
type ChunkedTransfer struct {
	chunkSize  int64
	compress   bool
	transferFn TransferFunc
	seen       map[string]bool
	mu         sync.Mutex
}

// NewChunkedTransfer creates a new ChunkedTransfer with the given config and transfer function.
func NewChunkedTransfer(cfg Config, fn TransferFunc) *ChunkedTransfer {
	cfg.Defaults()
	return &ChunkedTransfer{
		chunkSize:  cfg.ChunkSize,
		compress:   cfg.Compress,
		transferFn: fn,
		seen:       make(map[string]bool),
	}
}

// Transfer performs a deduplicated file transfer. If the file content (by SHA-256)
// has already been transferred successfully, the transfer is skipped.
// On failure, the hash is un-marked so that retries can re-attempt the transfer.
func (ct *ChunkedTransfer) Transfer(ctx context.Context, src, dst string) error {
	hash, err := ComputeHash(src)
	if err != nil {
		return fmt.Errorf("compute hash for %s: %w", src, err)
	}

	ct.mu.Lock()
	if ct.seen[hash] {
		ct.mu.Unlock()
		return nil // dedup: already transferred or in-flight
	}
	// Optimistically mark as seen to prevent concurrent duplicates.
	ct.seen[hash] = true
	ct.mu.Unlock()

	if err := ct.transferFn(ctx, src, dst); err != nil {
		// Un-mark on failure so retries can re-attempt.
		ct.mu.Lock()
		delete(ct.seen, hash)
		ct.mu.Unlock()
		return fmt.Errorf("transfer %s to %s: %w", src, dst, err)
	}
	return nil
}

// ComputeHash computes the SHA-256 hash of a file.
func ComputeHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file %s: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash file %s: %w", path, err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
