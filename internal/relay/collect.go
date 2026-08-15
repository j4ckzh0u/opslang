package relay

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// Collector handles fan-in file collection through a relay tree.
type Collector struct {
	config   Config
	topology *RelayNode
	transfer *ChunkedTransfer
}

// NewCollector creates a new Collector.
func NewCollector(cfg Config, topology *RelayNode, fn TransferFunc) *Collector {
	cfg.Defaults()
	return &Collector{
		config:   cfg,
		topology: topology,
		transfer: NewChunkedTransfer(cfg, fn),
	}
}

// Collect performs file collection from all leaf nodes to the destination directory.
// Each collected file is saved as {destDir}/{host}/{basename}.
func (c *Collector) Collect(ctx context.Context, task TransferTask, destDir string) (*CollectResult, error) {
	start := time.Now()

	result := &CollectResult{
		TaskID:  task.ID,
		DestDir: destDir,
	}

	leaves := collectLeaves(c.topology)
	result.Total = len(leaves)

	if len(leaves) == 0 {
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil
	}

	sem := make(chan struct{}, c.config.MaxConcurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, leaf := range leaves {
		wg.Add(1)
		go func(node *RelayNode) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:        // acquire
			case <-ctx.Done():
				ir := CollectItemResult{
					Host:   node.Host,
					Source: task.Source,
					Status: "failed",
					Error:  ctx.Err().Error(),
				}
				mu.Lock()
				result.Results = append(result.Results, ir)
				result.Failed++
				mu.Unlock()
				return
			}
			defer func() { <-sem }() // release

			hostDir := node.Host
			if node.HostInfo != nil {
				hostDir = node.HostInfo.Name
			}

			basename := filepath.Base(task.Source)
			localDest := filepath.Join(destDir, hostDir, basename)

			var ir CollectItemResult
			ir.Host = node.Host
			ir.Source = task.Source

			transferStart := time.Now()
			var lastErr error
			for attempt := 0; attempt <= c.config.Retries; attempt++ {
				if attempt > 0 {
					time.Sleep(time.Duration(attempt*100) * time.Millisecond)
				}

				// For collection, the source is on the remote and dest is local.
				// The TransferFunc is used in reverse: src=remote, dst=local.
				remoteSrc := fmt.Sprintf("%s:%s", node.Host, task.Source)
				err := c.transfer.Transfer(ctx, remoteSrc, localDest)
				if err == nil {
					ir.Status = "success"
					ir.Dest = localDest
					break
				}
				lastErr = err
			}

			if ir.Status != "success" {
				ir.Status = "failed"
				if lastErr != nil {
					ir.Error = lastErr.Error()
				}
			}
			ir.DurationMs = time.Since(transferStart).Milliseconds()

			mu.Lock()
			result.Results = append(result.Results, ir)
			switch ir.Status {
			case "success":
				result.Succeeded++
			case "failed":
				result.Failed++
			}
			mu.Unlock()
		}(leaf)
	}

	wg.Wait()
	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}
