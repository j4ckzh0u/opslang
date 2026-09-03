package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestDistributeRelaySwitchesCandidate(t *testing.T) {
	source := relayDistributionSource(t)
	targets := []DistributeTarget{
		{Host: "10.0.0.1", Dest: "/tmp/a", Tags: map[string]string{"relay": "true"}},
		{Host: "10.0.0.2", Dest: "/tmp/b"},
		{Host: "10.0.0.3", Dest: "/tmp/c"},
	}
	original := DefaultRelayGroupFunc
	t.Cleanup(func() { DefaultRelayGroupFunc = original })
	var candidates []string
	DefaultRelayGroupFunc = func(_ context.Context, _ string, relay DistributeTarget, peers []DistributeTarget, _ DistributeOptions) (map[string]HostDistributeResult, error) {
		candidates = append(candidates, relay.Host)
		if relay.Host == "10.0.0.1" {
			return nil, fmt.Errorf("candidate unavailable")
		}
		outcomes := map[string]HostDistributeResult{
			relayTargetIdentity(relay): {Host: relay.Host, Status: "success", Changed: true, TransferSource: "relay_seed"},
		}
		for _, peer := range peers {
			outcomes[relayTargetIdentity(peer)] = HostDistributeResult{Host: peer.Host, Status: "success", Changed: true, TransferSource: "relay:" + relay.Host}
		}
		return outcomes, nil
	}

	result, err := DistributeWith(source, targets, DistributeOptions{Relay: true, RelayThreshold: 2}, func(context.Context, string, string) error {
		return fmt.Errorf("direct fallback should not run")
	})
	if err != nil {
		t.Fatalf("distribute with relay: %v", err)
	}
	if !reflect.DeepEqual(candidates, []string{"10.0.0.1", "10.0.0.2"}) {
		t.Fatalf("candidates = %v", candidates)
	}
	if result.Succeeded != len(targets) || len(result.Results) != len(targets) {
		t.Fatalf("result = %+v", result)
	}
	for _, hostResult := range result.Results {
		if len(hostResult.Warnings) != 1 || !strings.Contains(hostResult.Warnings[0], "10.0.0.1") {
			t.Fatalf("candidate warning missing from %+v", hostResult)
		}
	}
}

func TestDistributeRelayFallsBackToDirectSFTP(t *testing.T) {
	source := relayDistributionSource(t)
	targets := []DistributeTarget{{Host: "10.0.0.1", Dest: "/tmp/a"}, {Host: "10.0.0.2", Dest: "/tmp/b"}}
	original := DefaultRelayGroupFunc
	t.Cleanup(func() { DefaultRelayGroupFunc = original })
	DefaultRelayGroupFunc = func(_ context.Context, _ string, relay DistributeTarget, _ []DistributeTarget, _ DistributeOptions) (map[string]HostDistributeResult, error) {
		return nil, fmt.Errorf("relay %s failed", relay.Host)
	}
	var mu sync.Mutex
	var endpoints []string
	result, err := DistributeWith(source, targets, DistributeOptions{Relay: true, RelayThreshold: 2, Parallel: 2, Retries: 1}, func(_ context.Context, _, endpoint string) error {
		mu.Lock()
		endpoints = append(endpoints, endpoint)
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("distribute fallback: %v", err)
	}
	if result.Succeeded != 2 || len(endpoints) != 2 {
		t.Fatalf("result = %+v, endpoints = %v", result, endpoints)
	}
	for _, hostResult := range result.Results {
		if hostResult.TransferSource != "direct_sftp" || len(hostResult.Warnings) != 2 {
			t.Fatalf("fallback result = %+v", hostResult)
		}
	}
}

func relayDistributionSource(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "payload.bin")
	if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return path
}
