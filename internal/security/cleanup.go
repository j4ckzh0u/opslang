package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TempDir manages a temporary directory with automatic cleanup
type TempDir struct {
	path    string
	cleaned bool
	mu      sync.Mutex
}

// NewTempDir creates a new temporary directory with the pattern /tmp/ops-<random>
func NewTempDir() (*TempDir, error) {
	// Generate random suffix
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random suffix: %w", err)
	}
	suffix := hex.EncodeToString(bytes)

	path := filepath.Join(os.TempDir(), fmt.Sprintf("ops-%s", suffix))

	// Create the directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	td := &TempDir{
		path:    path,
		cleaned: false,
	}

	// Register signal handler for cleanup
	td.registerSignalHandler()

	return td, nil
}

// Path returns the path to the temporary directory
func (td *TempDir) Path() string {
	return td.path
}

// Cleanup removes the temporary directory
func (td *TempDir) Cleanup() {
	td.mu.Lock()
	defer td.mu.Unlock()

	if td.cleaned {
		return
	}

	if err := os.RemoveAll(td.path); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to cleanup temp directory %s: %v\n", td.path, err)
	}

	td.cleaned = true
	globalTempDirs.Delete(td.path)
}

// package-level signal handler (registered once, shared by all TempDir instances)
var (
	globalSigOnce  sync.Once
	globalSigChan  = make(chan os.Signal, 1)
	globalTempDirs sync.Map // path -> *TempDir
)

// registerSignalHandler registers this TempDir with the package-level signal handler
func (td *TempDir) registerSignalHandler() {
	ensureGlobalSignalHandler()
	globalTempDirs.Store(td.path, td)
}

// IsCleaned returns whether the temp directory has been cleaned up
func (td *TempDir) IsCleaned() bool {
	td.mu.Lock()
	defer td.mu.Unlock()
	return td.cleaned
}
