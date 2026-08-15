package relay

import (
	"context"
	"time"
)

// RelayNode represents a node in the hierarchical relay tree.
type RelayNode struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	User     string            `json:"user"`
	Tier     int               `json:"tier"` // 0=controller, 1=L1 relay, 2=L2 relay
	Subnet   string            `json:"subnet"`
	Children []*RelayNode      `json:"children,omitempty"`
	Capacity int               `json:"capacity"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// IsLeaf indicates this node is a target (not a relay).
	IsLeaf bool `json:"is_leaf,omitempty"`
	// HostInfo stores the original host info for leaf nodes.
	HostInfo *HostInfo `json:"-"`
}

// HostInfo represents input host data from inventory.
type HostInfo struct {
	Name     string            `json:"name"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	User     string            `json:"user"`
	Password string            `json:"-"`
	KeyFile  string            `json:"key_file"`
	Tags     map[string]string `json:"tags"`
}

// TransferTask describes a file transfer operation.
type TransferTask struct {
	ID        string           `json:"id"`
	Source    string           `json:"source"`
	Dest      string           `json:"dest"`
	Compress  bool             `json:"compress"`
	Checksum  string           `json:"checksum"`
	ChunkSize int64            `json:"chunk_size"`
	Targets   []TransferTarget `json:"targets"`
}

// TransferTarget describes a single destination for a transfer.
type TransferTarget struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	DestPath string `json:"dest_path"`
}

// TransferResult is the result of transferring to a single host.
type TransferResult struct {
	Host       string `json:"host"`
	Status     string `json:"status"` // success/failed/skipped
	Changed    bool   `json:"changed"`
	Checksum   string `json:"checksum,omitempty"`
	Size       int64  `json:"size,omitempty"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// DistributeResult is the aggregate result of a distribute operation.
type DistributeResult struct {
	TaskID     string           `json:"task_id"`
	Total      int              `json:"total"`
	Succeeded  int              `json:"succeeded"`
	Failed     int              `json:"failed"`
	Skipped    int              `json:"skipped"`
	Results    []TransferResult `json:"results"`
	DurationMs int64            `json:"duration_ms"`
}

// CollectResult is the aggregate result of a collect operation.
type CollectResult struct {
	TaskID     string              `json:"task_id"`
	Total      int                 `json:"total"`
	Succeeded  int                 `json:"succeeded"`
	Failed     int                 `json:"failed"`
	Results    []CollectItemResult `json:"results"`
	DestDir    string              `json:"dest_dir"`
	DurationMs int64               `json:"duration_ms"`
}

// CollectItemResult is the result of collecting from a single host.
type CollectItemResult struct {
	Host       string `json:"host"`
	Status     string `json:"status"` // success/failed/skipped
	Source     string `json:"source"`
	Dest       string `json:"dest"`
	Checksum   string `json:"checksum,omitempty"`
	Size       int64  `json:"size,omitempty"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// Config holds configuration for relay operations.
type Config struct {
	MaxConcurrency   int           `json:"max_concurrency"`
	ChunkSize        int64         `json:"chunk_size"`
	Compress         bool          `json:"compress"`
	Retries          int           `json:"retries"`
	Timeout          time.Duration `json:"timeout"`
	RelayDepth       int           `json:"relay_depth"`
	MinHostsForRelay int           `json:"min_hosts_for_relay"`
}

// Defaults sets sensible default values for unset config fields.
func (c *Config) Defaults() {
	if c.MaxConcurrency <= 0 {
		c.MaxConcurrency = 10
	}
	if c.ChunkSize <= 0 {
		c.ChunkSize = 4 * 1024 * 1024 // 4MB
	}
	if c.Retries <= 0 {
		c.Retries = 3
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}
	if c.RelayDepth <= 0 {
		c.RelayDepth = 2
	}
	if c.MinHostsForRelay <= 0 {
		c.MinHostsForRelay = 10
	}
}

// TransferFunc is the function signature for actual file transfer operations.
// Implementations handle the real SSH/SFTP transfer; tests use mocks.
type TransferFunc func(ctx context.Context, src, dst string) error
