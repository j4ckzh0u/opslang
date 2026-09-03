package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	opsfile "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/file"
)

func TestRunRelayServeIsBounded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.bin")
	if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	var output bytes.Buffer
	if err := runRelay(context.Background(), []string{"serve", "--file", path, "--ttl", "20ms"}, &output); err != nil {
		t.Fatalf("run relay serve: %v", err)
	}
	var info opsfile.RelayHTTPInfo
	if err := json.Unmarshal(output.Bytes(), &info); err != nil {
		t.Fatalf("decode relay info: %v", err)
	}
	if info.URL == "" || info.Token == "" || info.CertFingerprint == "" {
		t.Fatalf("incomplete relay info: %+v", info)
	}
}

func TestRunRelayFetch(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.bin")
	content := []byte("relay runner payload")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	server, err := opsfile.StartRelayHTTPServer(source, time.Minute, 1)
	if err != nil {
		t.Fatalf("start relay server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if warnings := server.Stop(ctx); len(warnings) > 0 {
			t.Errorf("stop warnings: %v", warnings)
		}
	}()
	destination := filepath.Join(t.TempDir(), "destination.bin")
	var output bytes.Buffer
	args := []string{
		"fetch",
		"--url", server.Info.URL,
		"--token", server.Info.Token,
		"--fingerprint", server.Info.CertFingerprint,
		"--sha256", server.Info.SHA256,
		"--size", strconv.FormatInt(server.Info.Size, 10),
		"--dest", destination,
	}
	if err := runRelay(context.Background(), args, &output); err != nil {
		t.Fatalf("run relay fetch: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("destination = %q, want %q", got, content)
	}
	var outcome opsfile.TransferOutcome
	if err := json.Unmarshal(output.Bytes(), &outcome); err != nil {
		t.Fatalf("decode outcome: %v", err)
	}
	if outcome.Status != "success" {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestRunRelayValidatesSubcommand(t *testing.T) {
	if err := runRelay(context.Background(), nil, &bytes.Buffer{}); err == nil {
		t.Fatal("expected missing subcommand error")
	}
	if err := runRelay(context.Background(), []string{"unknown"}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected unknown subcommand error")
	}
}
