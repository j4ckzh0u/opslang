package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
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
		// Log error but don't fail - this is cleanup
		fmt.Fprintf(os.Stderr, "warning: failed to cleanup temp directory %s: %v\n", td.path, err)
	}

	td.cleaned = true
}

// registerSignalHandler registers SIGINT and SIGTERM handlers for cleanup
func (td *TempDir) registerSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		td.Cleanup()
		// Re-raise signal to exit
		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
		syscall.Kill(os.Getpid(), syscall.SIGINT)
	}()
}

// IsCleaned returns whether the temp directory has been cleaned up
func (td *TempDir) IsCleaned() bool {
	td.mu.Lock()
	defer td.mu.Unlock()
	return td.cleaned
}
