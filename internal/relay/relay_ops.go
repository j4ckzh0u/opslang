package relay

import (
	"context"
	"fmt"
	"time"

	"github.com/opslang/opslang/internal/runner"
)

// RegisterOperations registers relay operations (file.distribute, file.collect)
// with the given runner registry.
func RegisterOperations(reg *runner.Registry) {
	reg.Register("file.distribute", distributeOp)
	reg.Register("file.collect", collectOp)
}

// distributeOp handles the file.distribute operation from the runner.
func distributeOp(args map[string]interface{}) (interface{}, error) {
	source := requireStringArg(args, "source")
	dest := requireStringArg(args, "dest")
	if source == "" || dest == "" {
		return nil, fmt.Errorf("source and dest are required")
	}

	compress, _ := args["compress"].(bool)

	targets := parseTargets(args)

	// Build host infos from targets.
	hosts := make([]HostInfo, len(targets))
	for i, t := range targets {
		hosts[i] = HostInfo{
			Name: t.Host,
			Host: t.Host,
			Port: t.Port,
			User: t.User,
		}
	}

	cfg := Config{}
	cfg.Defaults()

	topology, err := BuildTopology(hosts, cfg)
	if err != nil {
		return nil, fmt.Errorf("build topology: %w", err)
	}

	// Placeholder transfer function (real implementation would use SSH/SFTP).
	transferFn := func(_ context.Context, _, _ string) error {
		return nil // placeholder
	}

	d := NewDistributor(cfg, topology, transferFn)
	task := TransferTask{
		ID:       fmt.Sprintf("distribute-%d", time.Now().UnixNano()),
		Source:   source,
		Dest:     dest,
		Compress: compress,
	}

	return d.Distribute(context.Background(), task)
}

// collectOp handles the file.collect operation from the runner.
func collectOp(args map[string]interface{}) (interface{}, error) {
	source := requireStringArg(args, "source")
	dest := requireStringArg(args, "dest")
	if source == "" || dest == "" {
		return nil, fmt.Errorf("source and dest are required")
	}

	targets := parseTargets(args)

	hosts := make([]HostInfo, len(targets))
	for i, t := range targets {
		hosts[i] = HostInfo{
			Name: t.Host,
			Host: t.Host,
			Port: t.Port,
			User: t.User,
		}
	}

	cfg := Config{}
	cfg.Defaults()

	topology, err := BuildTopology(hosts, cfg)
	if err != nil {
		return nil, fmt.Errorf("build topology: %w", err)
	}

	transferFn := func(_ context.Context, _, _ string) error {
		return nil // placeholder
	}

	c := NewCollector(cfg, topology, transferFn)
	task := TransferTask{
		ID:     fmt.Sprintf("collect-%d", time.Now().UnixNano()),
		Source: source,
		Dest:   dest,
	}

	return c.Collect(context.Background(), task, dest)
}

// parseTargets extracts transfer targets from the args map.
func parseTargets(args map[string]interface{}) []TransferTarget {
	var targets []TransferTarget
	if t, ok := args["targets"].([]interface{}); ok {
		for _, item := range t {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			target := TransferTarget{}
			if h, ok := m["host"].(string); ok {
				target.Host = h
			}
			if p, ok := m["port"].(float64); ok {
				target.Port = int(p)
			}
			if u, ok := m["user"].(string); ok {
				target.User = u
			}
			if dp, ok := m["dest_path"].(string); ok {
				target.DestPath = dp
			}
			targets = append(targets, target)
		}
	}
	return targets
}

// requireStringArg extracts a string argument from the args map.
func requireStringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
