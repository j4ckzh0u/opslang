package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheNew(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestCacheNewDefault(t *testing.T) {
	c, err := NewCache("")
	if err != nil {
		t.Fatalf("NewCache with empty dir failed: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestCacheKey(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	key1 := c.Key("source1", "linux/amd64")
	key2 := c.Key("source2", "linux/amd64")
	key3 := c.Key("source1", "linux/arm64")

	if key1 == key2 {
		t.Error("different sources should produce different keys")
	}
	if key1 == key3 {
		t.Error("different architectures should produce different keys")
	}
	if len(key1) != 32 {
		t.Errorf("expected key length 32, got %d", len(key1))
	}
}

func TestCacheKeyDeterministic(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	key1 := c.Key("same source", "linux/amd64")
	key2 := c.Key("same source", "linux/amd64")

	if key1 != key2 {
		t.Error("same inputs should produce same key")
	}
}

func TestCacheGetMiss(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	result := c.Get("nonexistent-key")
	if result != "" {
		t.Errorf("expected empty string for cache miss, got %q", result)
	}
}

func TestCachePutAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	// Create a fake binary
	fakeBinary := filepath.Join(tmpDir, "fake-binary")
	if err := os.WriteFile(fakeBinary, []byte("fake-binary-content"), 0755); err != nil {
		t.Fatalf("failed to write fake binary: %v", err)
	}

	key := c.Key("test-source", "linux/amd64")

	// Put into cache
	if err := c.Put(key, fakeBinary); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get from cache
	cachedPath := c.Get(key)
	if cachedPath == "" {
		t.Fatal("expected cache hit, got empty string")
	}

	// Verify content
	data, err := os.ReadFile(cachedPath)
	if err != nil {
		t.Fatalf("failed to read cached file: %v", err)
	}
	if string(data) != "fake-binary-content" {
		t.Errorf("expected 'fake-binary-content', got %q", string(data))
	}
}

func TestCacheDifferentSourcesMiss(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	// Create and cache a binary
	fakeBinary := filepath.Join(tmpDir, "fake-binary")
	os.WriteFile(fakeBinary, []byte("content"), 0755)

	key1 := c.Key("source1", "linux/amd64")
	c.Put(key1, fakeBinary)

	// Different source should miss
	key2 := c.Key("source2", "linux/amd64")
	result := c.Get(key2)
	if result != "" {
		t.Error("different source should not hit cache")
	}
}
