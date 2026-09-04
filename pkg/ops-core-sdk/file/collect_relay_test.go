package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCollectRelaySwitchesCandidate(t *testing.T) {
	destDir := t.TempDir()
	targets := []CollectTarget{
		{Host: "10.0.0.1", Source: "/var/log/app.log", Tags: map[string]string{"relay": "true"}},
		{Host: "10.0.0.2", Source: "/var/log/app.log"},
		{Host: "10.0.0.3", Source: "/var/log/app.log"},
	}
	original := DefaultCollectRelayGroupFunc
	t.Cleanup(func() { DefaultCollectRelayGroupFunc = original })
	var candidates []string
	DefaultCollectRelayGroupFunc = func(_ context.Context, _ string, relay CollectTarget, peers []CollectTarget, _ CollectOptions) (map[string]HostCollectResult, error) {
		candidates = append(candidates, relay.Host)
		if relay.Host == "10.0.0.1" {
			return nil, fmt.Errorf("candidate unavailable")
		}
		outcomes := map[string]HostCollectResult{
			collectTargetIdentity(relay): {Host: relay.Host, Status: "success", TransferSource: "relay:" + relay.Host},
		}
		for _, peer := range peers {
			outcomes[collectTargetIdentity(peer)] = HostCollectResult{Host: peer.Host, Status: "success", TransferSource: "relay:" + relay.Host}
		}
		return outcomes, nil
	}

	result, err := CollectWith("/var/log/app.log", targets, CollectOptions{DestDir: destDir, Relay: true, RelayThreshold: 2}, func(context.Context, string, string) error {
		return fmt.Errorf("direct fallback should not run")
	})
	if err != nil {
		t.Fatalf("collect with relay: %v", err)
	}
	if !reflect.DeepEqual(candidates, []string{"10.0.0.1", "10.0.0.2"}) {
		t.Fatalf("candidates = %v", candidates)
	}
	if result.Succeeded != len(targets) || len(result.Results) != len(targets) {
		t.Fatalf("result = %+v", result)
	}
	for _, hostResult := range result.Results {
		if hostResult.TransferSource != "relay:10.0.0.2" || len(hostResult.Warnings) != 1 || !strings.Contains(hostResult.Warnings[0], "10.0.0.1") {
			t.Fatalf("relay result = %+v", hostResult)
		}
	}
}

func TestCollectRelayFallsBackToDirect(t *testing.T) {
	destDir := t.TempDir()
	targets := []CollectTarget{{Host: "10.0.0.1"}, {Host: "10.0.0.2"}}
	original := DefaultCollectRelayGroupFunc
	t.Cleanup(func() { DefaultCollectRelayGroupFunc = original })
	DefaultCollectRelayGroupFunc = func(_ context.Context, _ string, relay CollectTarget, _ []CollectTarget, _ CollectOptions) (map[string]HostCollectResult, error) {
		return nil, fmt.Errorf("relay %s failed", relay.Host)
	}
	var endpoints []string
	result, err := CollectWith("/tmp/source", targets, CollectOptions{DestDir: destDir, Relay: true, RelayThreshold: 2, Retries: 1}, func(_ context.Context, source, destination string) error {
		endpoints = append(endpoints, source)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		return os.WriteFile(destination, []byte("data"), 0o600)
	})
	if err != nil {
		t.Fatalf("collect fallback: %v", err)
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
