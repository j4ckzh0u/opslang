package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CollectRelayGroupFunc collects one source group through a relay host.
// The relay implementation makes each source host serve a short-lived HTTPS
// session, asks the relay to pull it, then downloads the relay copy.
type CollectRelayGroupFunc func(ctx context.Context, source string, relay CollectTarget, targets []CollectTarget, opts CollectOptions) (map[string]HostCollectResult, error)

// DefaultCollectRelayGroupFunc is configured by WireSSHTransfer.
var DefaultCollectRelayGroupFunc CollectRelayGroupFunc = func(_ context.Context, _ string, _ CollectTarget, _ []CollectTarget, _ CollectOptions) (map[string]HostCollectResult, error) {
	return nil, fmt.Errorf("no relay collect function configured")
}

func collectWithRelay(source string, targets []CollectTarget, opts CollectOptions, downloadFn CollectDownloadFunc) (*CollectResult, error) {
	started := time.Now()
	distributeTargets := make([]DistributeTarget, len(targets))
	for index, target := range targets {
		distributeTargets[index] = DistributeTarget{Host: target.Host, Port: target.Port, User: target.User, RelayGroup: target.RelayGroup, Tags: target.Tags}
	}
	plan, err := BuildRelayPlan(distributeTargets, DistributeOptions{
		Relay: true, RelayGroup: opts.RelayGroup, RelayThreshold: opts.RelayThreshold, RelayMaxTargets: opts.RelayMaxTargets,
	})
	if err != nil {
		return nil, fmt.Errorf("file.Collect: build relay plan: %w", err)
	}
	byIdentity := make(map[string]HostCollectResult, len(targets))
	fallback := make([]CollectTarget, 0, len(targets))
	warnings := make(map[string][]string, len(targets))
	targetByRelayIdentity := make(map[string]CollectTarget, len(targets))
	for _, target := range targets {
		targetByRelayIdentity[collectTargetRelayIdentity(target)] = target
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	for _, group := range plan.Groups {
		if group.Relay == nil {
			for _, target := range group.Direct {
				fallback = append(fallback, targetByRelayIdentity[relayTargetIdentity(target)])
			}
			continue
		}
		members := append([]DistributeTarget{*group.Relay}, group.Targets...)
		completedGroup := false
		for candidateIndex, candidate := range members {
			relay := targetByRelayIdentity[relayTargetIdentity(candidate)]
			peers := make([]CollectTarget, 0, len(members)-1)
			for index, member := range members {
				if index != candidateIndex {
					peers = append(peers, targetByRelayIdentity[relayTargetIdentity(member)])
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			outcomes, relayErr := DefaultCollectRelayGroupFunc(ctx, source, relay, peers, opts)
			cancel()
			if relayErr != nil {
				message := fmt.Sprintf("relay candidate %s failed: %v", candidate.Host, relayErr)
				for _, member := range members {
					identity := collectTargetIdentity(targetByRelayIdentity[relayTargetIdentity(member)])
					warnings[identity] = append(warnings[identity], message)
				}
				continue
			}
			for _, member := range members {
				identity := collectTargetIdentity(targetByRelayIdentity[relayTargetIdentity(member)])
				outcome, ok := outcomes[identity]
				if ok && (outcome.Status == "success" || outcome.Status == "skipped") {
					outcome.Warnings = append(warnings[identity], outcome.Warnings...)
					byIdentity[identity] = outcome
					continue
				}
				if ok && outcome.Error != "" {
					warnings[identity] = append(warnings[identity], "relay transfer failed: "+outcome.Error)
				}
				fallback = append(fallback, targetByRelayIdentity[relayTargetIdentity(member)])
			}
			completedGroup = true
			break
		}
		if !completedGroup {
			for _, member := range members {
				fallback = append(fallback, targetByRelayIdentity[relayTargetIdentity(member)])
			}
		}
	}
	directOpts := opts
	directOpts.Relay = false
	directResult, err := CollectWith(source, fallback, directOpts, downloadFn)
	if err != nil {
		return nil, err
	}
	directOutcomes := make(map[string]HostCollectResult, len(directResult.Results))
	for _, outcome := range directResult.Results {
		directOutcomes[outcome.Host+":"+outcome.Source] = outcome
	}
	for _, target := range fallback {
		outcome, ok := directOutcomes[target.Host+":"+target.Source]
		if !ok {
			return nil, fmt.Errorf("file.Collect: direct fallback result missing for %s", target.Host)
		}
		identity := collectTargetIdentity(target)
		outcome.Warnings = append(warnings[identity], outcome.Warnings...)
		if outcome.TransferSource == "" {
			outcome.TransferSource = "direct_sftp"
		}
		byIdentity[identity] = outcome
	}
	result := &CollectResult{Total: len(targets), Results: make([]HostCollectResult, 0, len(targets))}
	for _, target := range targets {
		outcome, ok := byIdentity[collectTargetIdentity(target)]
		if !ok {
			return nil, fmt.Errorf("file.Collect: relay result missing for %s", target.Host)
		}
		result.Results = append(result.Results, outcome)
		switch outcome.Status {
		case "success":
			result.Succeeded++
		case "skipped":
			result.Skipped++
		default:
			result.Failed++
		}
	}
	result.DestDir = opts.DestDir
	if result.DestDir == "" {
		result.DestDir = filepath.Join(os.TempDir(), "ops-collect")
	}
	result.DurationMs = time.Since(started).Milliseconds()
	return result, nil
}

func collectTargetIdentity(target CollectTarget) string {
	return target.Host + ":" + fmt.Sprint(target.Port) + ":" + target.User + ":" + target.Source
}

func collectTargetRelayIdentity(target CollectTarget) string {
	return target.Host + ":" + fmt.Sprint(target.Port) + ":" + target.User + ":"
}

func collectDestination(destDir string, target CollectTarget, fallbackSource string) string {
	if destDir == "" {
		destDir = filepath.Join(os.TempDir(), "ops-collect")
	}
	source := target.Source
	if source == "" {
		source = fallbackSource
	}
	host := target.Host
	if host == "" {
		host = "unknown"
	}
	return filepath.Join(destDir, host, filepath.Base(source))
}
