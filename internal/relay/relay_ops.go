package relay

import (
	"fmt"
	"time"

	sdkfile "github.com/opslang/opslang/pkg/ops-core-sdk/file"
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
	if source == "" {
		return nil, fmt.Errorf("file.distribute: source is required")
	}

	targets := parseDistributeTargets(args)
	if len(targets) == 0 {
		return nil, fmt.Errorf("file.distribute: at least one target is required")
	}

	opts := sdkfile.DistributeOptions{
		Compress: getBoolArg(args, "compress"),
		Checksum: getBoolArg(args, "checksum"),
		Mode:     requireStringArg(args, "mode"),
		Owner:    requireStringArg(args, "owner"),
		Parallel: getIntArg(args, "parallel"),
		Retries:  getIntArg(args, "retries"),
	}
	if t := getIntArg(args, "timeout_ms"); t > 0 {
		opts.Timeout = time.Duration(t) * time.Millisecond
	}

	return sdkfile.Distribute(source, targets, opts)
}

// collectOp handles the file.collect operation from the runner.
func collectOp(args map[string]interface{}) (interface{}, error) {
	source := requireStringArg(args, "source")
	if source == "" {
		return nil, fmt.Errorf("file.collect: source is required")
	}

	dest := requireStringArg(args, "dest")
	if dest == "" {
		return nil, fmt.Errorf("file.collect: dest is required")
	}

	targets := parseCollectTargets(args, source)
	if len(targets) == 0 {
		return nil, fmt.Errorf("file.collect: at least one target is required")
	}

	opts := sdkfile.CollectOptions{
		Compress: getBoolArg(args, "compress"),
		DestDir:  dest,
		Parallel: getIntArg(args, "parallel"),
		Retries:  getIntArg(args, "retries"),
	}
	if t := getIntArg(args, "timeout_ms"); t > 0 {
		opts.Timeout = time.Duration(t) * time.Millisecond
	}

	return sdkfile.Collect(source, targets, opts)
}

// parseDistributeTargets extracts distribute targets from the args map.
func parseDistributeTargets(args map[string]interface{}) []sdkfile.DistributeTarget {
	var targets []sdkfile.DistributeTarget
	defaultDest := requireStringArg(args, "dest")

	if t, ok := args["targets"].([]interface{}); ok {
		for _, item := range t {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			target := sdkfile.DistributeTarget{
				Dest: defaultDest,
			}
			if h, ok := m["host"].(string); ok {
				target.Host = h
			}
			if p, ok := m["port"].(float64); ok {
				target.Port = int(p)
			}
			if u, ok := m["user"].(string); ok {
				target.User = u
			}
			if d, ok := m["dest"].(string); ok && d != "" {
				target.Dest = d
			}
			if d, ok := m["dest_path"].(string); ok && d != "" {
				target.Dest = d
			}
			targets = append(targets, target)
		}
	}
	return targets
}

// parseCollectTargets extracts collect targets from the args map.
func parseCollectTargets(args map[string]interface{}, defaultSource string) []sdkfile.CollectTarget {
	var targets []sdkfile.CollectTarget

	if t, ok := args["targets"].([]interface{}); ok {
		for _, item := range t {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			target := sdkfile.CollectTarget{
				Source: defaultSource,
			}
			if h, ok := m["host"].(string); ok {
				target.Host = h
			}
			if p, ok := m["port"].(float64); ok {
				target.Port = int(p)
			}
			if u, ok := m["user"].(string); ok {
				target.User = u
			}
			if s, ok := m["source"].(string); ok && s != "" {
				target.Source = s
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

// getBoolArg extracts a bool argument from the args map.
func getBoolArg(args map[string]interface{}, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// getIntArg extracts an int argument from the args map (supports float64 from JSON).
func getIntArg(args map[string]interface{}, key string) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case int64:
			return int(n)
		}
	}
	return 0
}
