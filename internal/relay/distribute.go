package relay

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Distributor handles fan-out file distribution through a relay tree.
type Distributor struct {
	config   Config
	topology *RelayNode
	transfer *ChunkedTransfer
}

// NewDistributor creates a new Distributor.
func NewDistributor(cfg Config, topology *RelayNode, fn TransferFunc) *Distributor {
	cfg.Defaults()
	return &Distributor{
		config:   cfg,
		topology: topology,
		transfer: NewChunkedTransfer(cfg, fn),
	}
}

// Distribute performs the file distribution across the relay tree.
// All leaf nodes receive the file; concurrency is bounded by config.MaxConcurrency.
func (d *Distributor) Distribute(ctx context.Context, task TransferTask) (*DistributeResult, error) {
	start := time.Now()

	result := &DistributeResult{
		TaskID: task.ID,
	}

	// Collect all leaf nodes.
	leaves := collectLeaves(d.topology)
	result.Total = len(leaves)

	if len(leaves) == 0 {
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil
	}

	// Use a semaphore for concurrency control.
	sem := make(chan struct{}, d.config.MaxConcurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, leaf := range leaves {
		wg.Add(1)
		go func(node *RelayNode) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:        // acquire
			case <-ctx.Done():
				tr := TransferResult{
					Host:   node.Host,
					Status: "failed",
					Error:  ctx.Err().Error(),
				}
				mu.Lock()
				result.Results = append(result.Results, tr)
				result.Failed++
				mu.Unlock()
				return
			}
			defer func() { <-sem }() // release

			destPath := task.Dest
			if node.IsLeaf && node.HostInfo != nil {
				destPath = fmt.Sprintf("%s/%s", task.Dest, node.HostInfo.Name)
			}

			var tr TransferResult
			tr.Host = node.Host

			transferStart := time.Now()
			var lastErr error
			for attempt := 0; attempt <= d.config.Retries; attempt++ {
				if attempt > 0 {
					// Brief backoff.
					time.Sleep(time.Duration(attempt*100) * time.Millisecond)
				}

				err := d.transfer.Transfer(ctx, task.Source, destPath)
				if err == nil {
					tr.Status = "success"
					tr.Changed = true
					tr.DurationMs = time.Since(transferStart).Milliseconds()
					break
				}
				lastErr = err
			}

			if tr.Status != "success" {
				tr.Status = "failed"
				tr.DurationMs = time.Since(transferStart).Milliseconds()
				if lastErr != nil {
					tr.Error = lastErr.Error()
				}
			}

			mu.Lock()
			result.Results = append(result.Results, tr)
			switch tr.Status {
			case "success":
				result.Succeeded++
			case "failed":
				result.Failed++
			case "skipped":
				result.Skipped++
			}
			mu.Unlock()
		}(leaf)
	}

	wg.Wait()
	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

// collectLeaves recursively collects all leaf nodes from a relay tree.
func collectLeaves(node *RelayNode) []*RelayNode {
	if node == nil {
		return nil
	}
	if node.IsLeaf {
		return []*RelayNode{node}
	}
	var leaves []*RelayNode
	for _, child := range node.Children {
		leaves = append(leaves, collectLeaves(child)...)
	}
	return leaves
}
