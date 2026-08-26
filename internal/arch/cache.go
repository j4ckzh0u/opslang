// arch/cache.go: two-tier cache for remote architecture detection.
//
// A host's CPU architecture effectively never changes, but without a cache
// every deploy pays one extra SSH round-trip (`uname -m`) per host. This
// cache keeps results in process for the current run and optionally on disk
// so later runs skip the probe too. Entries carry a TTL because "effectively
// never" is not "never" (VMs can be migrated across architectures).
//
// Failure policy: the cache may never break a deploy. An unreadable or
// corrupt cache file is ignored and overwritten on the next save; detection
// errors are never cached.
package arch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DefaultTTL is how long a cached architecture stays trusted. One day is
// generous for hardware that cannot change under a running OS.
const DefaultTTL = 24 * time.Hour

// entry is one cached detection result.
type entry struct {
	GOARCH   string    `json:"goarch"`
	CachedAt time.Time `json:"cached_at"`
}

// Cache stores detected architectures keyed by "host:port".
type Cache struct {
	mu  sync.Mutex
	mem map[string]entry
	// path is the backing file; empty disables disk persistence.
	path string
	ttl  time.Duration
}

// NewCache creates a cache backed by path ("" means memory only).
// ttl <= 0 means entries never expire.
func NewCache(path string, ttl time.Duration) *Cache {
	return &Cache{
		mem:  make(map[string]entry),
		path: path,
		ttl:  ttl,
	}
}

// DefaultCachePath returns ~/.opsctl/arch-cache.json.
func DefaultCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".opsctl", "arch-cache.json"), nil
}

// Load reads the disk cache into memory. Missing or corrupt files are
// tolerated silently at this level: the caller's deploy must proceed, and
// the next Save rewrites the file. The error is reported only so callers
// that care (tests, verbose logs) can surface it.
func (c *Cache) Load() error {
	if c.path == "" {
		return nil
	}
	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read arch cache: %w", err)
	}
	var disk map[string]entry
	if err := json.Unmarshal(data, &disk); err != nil {
		return fmt.Errorf("arch cache corrupted, ignoring: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range disk {
		if _, ok := c.mem[k]; !ok {
			c.mem[k] = v
		}
	}
	return nil
}

// Get returns the cached GOARCH for targetID when a fresh entry exists.
func (c *Cache) Get(targetID string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.mem[targetID]
	if !ok {
		return "", false
	}
	if c.ttl > 0 && time.Since(e.CachedAt) > c.ttl {
		return "", false
	}
	return e.GOARCH, true
}

// Put records a detection result in memory and persists to disk.
// Disk write failures do not invalidate the in-memory value: the current
// run still benefits; only future runs fall back to re-detection.
func (c *Cache) Put(targetID, goarch string) error {
	c.mu.Lock()
	c.mem[targetID] = entry{GOARCH: goarch, CachedAt: time.Now().UTC()}
	c.mu.Unlock()
	return c.save()
}

// Detect returns the architecture for targetID, consulting the cache first
// and falling back to a live `uname -m` probe whose result is cached.
func (c *Cache) Detect(ctx context.Context, executor SSHExecutor, targetID string) (string, error) {
	if goarch, ok := c.Get(targetID); ok {
		return goarch, nil
	}
	goarch, err := Detect(ctx, executor)
	if err != nil {
		// Never cache failures: a transient SSH error must not pin a
		// host to "undetectable" for the TTL window.
		return "", err
	}
	if perr := c.Put(targetID, goarch); perr != nil {
		// Detection succeeded; persistence problems must not fail the host.
		_ = perr
	}
	return goarch, nil
}

// save atomically persists the in-memory map: write a temp file in the same
// directory, then rename over the target so concurrent readers never see a
// half-written cache.
func (c *Cache) save() error {
	if c.path == "" {
		return nil
	}
	c.mu.Lock()
	snapshot := make(map[string]entry, len(c.mem))
	for k, v := range c.mem {
		snapshot[k] = v
	}
	c.mu.Unlock()

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal arch cache: %w", err)
	}
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create arch cache dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".arch-cache-*")
	if err != nil {
		return fmt.Errorf("failed to create arch cache temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after successful rename

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("failed to write arch cache: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close arch cache temp file: %w", err)
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		return fmt.Errorf("failed to chmod arch cache: %w", err)
	}
	if err := os.Rename(tmpName, c.path); err != nil {
		return fmt.Errorf("failed to replace arch cache: %w", err)
	}
	return nil
}
