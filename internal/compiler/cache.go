package compiler

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// Cache manages compiled binary caching.
type Cache struct {
	dir string // cache directory, defaults to ~/.opsctl/cache
}

// NewCache creates a new cache in the given directory.
// If dir is empty, uses ~/.opsctl/cache.
func NewCache(dir string) (*Cache, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = filepath.Join(home, ".opsctl", "cache")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{dir: dir}, nil
}

// Key computes a cache key from source content and target architecture.
func (c *Cache) Key(source string, targetArch string) string {
	h := sha256.Sum256([]byte(source + "|" + targetArch))
	return fmt.Sprintf("%x", h[:16]) // 32 hex chars
}

// Get looks up a cached binary by key. Returns the path or empty string if not found.
func (c *Cache) Get(key string) string {
	path := filepath.Join(c.dir, key)
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path
	}
	return ""
}

// Put stores a compiled binary in the cache under the given key.
func (c *Cache) Put(key string, binaryPath string) error {
	destPath := filepath.Join(c.dir, key)

	data, err := os.ReadFile(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to read binary for caching: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0755); err != nil {
		return fmt.Errorf("failed to write cached binary: %w", err)
	}

	return nil
}
