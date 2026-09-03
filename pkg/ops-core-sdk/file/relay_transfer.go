package file

import (
	"context"
	"fmt"
	"time"
)

// RelayGroupFunc transfers one source through a selected relay candidate.
type RelayGroupFunc func(ctx context.Context, source string, relay DistributeTarget, targets []DistributeTarget, opts DistributeOptions) (map[string]HostDistributeResult, error)

// DefaultRelayGroupFunc is configured by the SSH transport implementation.
var DefaultRelayGroupFunc RelayGroupFunc = func(_ context.Context, _ string, _ DistributeTarget, _ []DistributeTarget, _ DistributeOptions) (map[string]HostDistributeResult, error) {
	return nil, fmt.Errorf("no relay group transfer function configured")
}

func distributeWithRelay(source string, targets []DistributeTarget, opts DistributeOptions, transferFn TransferFunc) (*DistributeResult, error) {
	started := time.Now()
	plan, err := BuildRelayPlan(targets, opts)
	if err != nil {
		return nil, fmt.Errorf("file.Distribute: build relay plan: %w", err)
	}
	completed := make(map[string]HostDistributeResult, len(targets))
	fallback := make([]DistributeTarget, 0, len(targets))
	warnings := make(map[string][]string, len(targets))
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	for _, group := range plan.Groups {
		if group.Relay == nil {
			fallback = append(fallback, group.Direct...)
			continue
		}
		members := append([]DistributeTarget{cloneDistributeTarget(*group.Relay)}, group.Targets...)
		groupCompleted := false
		for candidateIndex, candidate := range members {
			peers := make([]DistributeTarget, 0, len(members)-1)
			peers = append(peers, members[:candidateIndex]...)
			peers = append(peers, members[candidateIndex+1:]...)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			outcomes, relayErr := DefaultRelayGroupFunc(ctx, source, candidate, peers, opts)
			cancel()
			if relayErr != nil {
				message := fmt.Sprintf("relay candidate %s failed: %v", candidate.Host, relayErr)
				for _, member := range members {
					identity := relayTargetIdentity(member)
					warnings[identity] = append(warnings[identity], message)
				}
				continue
			}
			for _, member := range members {
				identity := relayTargetIdentity(member)
				outcome, ok := outcomes[identity]
				if ok && (outcome.Status == "success" || outcome.Status == "skipped") {
					outcome.Warnings = append(warnings[identity], outcome.Warnings...)
					completed[identity] = outcome
					continue
				}
				if ok && outcome.Error != "" {
					warnings[identity] = append(warnings[identity], "relay transfer failed: "+outcome.Error)
				}
				fallback = append(fallback, member)
			}
			groupCompleted = true
			break
		}
		if !groupCompleted {
			fallback = append(fallback, members...)
		}
	}

	directOptions := opts
	directOptions.Relay = false
	directResult, err := DistributeWith(source, fallback, directOptions, transferFn)
	if err != nil {
		return nil, err
	}
	for index, target := range fallback {
		outcome := directResult.Results[index]
		identity := relayTargetIdentity(target)
		outcome.Warnings = append(warnings[identity], outcome.Warnings...)
		if outcome.TransferSource == "" {
			outcome.TransferSource = "direct_sftp"
		}
		completed[identity] = outcome
	}

	result := &DistributeResult{Total: len(targets), Results: make([]HostDistributeResult, 0, len(targets))}
	for _, target := range targets {
		outcome, ok := completed[relayTargetIdentity(target)]
		if !ok {
			return nil, fmt.Errorf("file.Distribute: relay result missing for %s", target.Host)
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
	result.DurationMs = time.Since(started).Milliseconds()
	return result, nil
}
